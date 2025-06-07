package app

import (
	"api/gen/api/v1/apiv1connect"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"api/app/domain/services"
	"api/app/domain/usecases"
	"api/app/internal/config"
	"api/app/internal/presentation"

	"api/app/internal/infrastructure/external/card_craft_ai"
	"api/app/internal/infrastructure/external/ozon"
	"api/app/internal/infrastructure/external/token_counter"
	"api/app/internal/infrastructure/external/wb"
	pgstorage "api/app/internal/infrastructure/persistence/postgres"

	"api/metrics"

	"github.com/marketconnect/db_client/postgresql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	cardCraftAiAPIURL string
	httpClient        *http.Client
	cfg               *config.Config
	mux               *http.ServeMux
	balanceStorage    *pgstorage.BalanceStorage
}

// NewApp creates a new ProductServer instance
func NewApp() *App {

	cfg := config.GetConfig()

	pgCfg := postgresql.NewPgConfig(
		cfg.PostgreSQL.Username,
		cfg.PostgreSQL.Password,
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.Port,
		cfg.PostgreSQL.Database,
	)
	pgClient, err := postgresql.NewClient(context.Background(), 3, time.Second, pgCfg)
	if err != nil {
		log.Fatalf("failed to init postgres client: %v", err)
	}
	balanceStorage := pgstorage.NewBalanceStorage(pgClient)

	// clients
	cardCraftAiClient := card_craft_ai.NewCardCraftAiClient(cfg.CardCraftAi.APIURL)
	wbClient := wb.NewWBClient()
	ozonClient := ozon.NewClient()
	tokenCounterClient := token_counter.NewClient(cfg.TokenCounter.APIURL)

	// services
	cardCraftAiService := services.NewCardCraftAiService(cardCraftAiClient)
	tokenBillingService := services.NewTokenBillingService(tokenCounterClient, balanceStorage)
	wbService := services.NewWbService(cfg.WB.GetCardListMaxAttempts, wbClient)
	ozonService := services.NewOzonService(ozonClient)

	// usecases
	createCardUsecase := usecases.NewCreateCardUsecase(cardCraftAiService, wbService, ozonService, tokenBillingService)
	getBalanceUsecase := usecases.NewGetBalanceUsecase(balanceStorage)

	// handlers
	createProductCardHandler := presentation.NewCreateProductCardHandler(createCardUsecase)
	balanceHandler := presentation.NewBalanceHandler(getBalanceUsecase)

	// mux
	mux := http.NewServeMux()
	path, baseHandler := apiv1connect.NewCreateProductCardServiceHandler(createProductCardHandler)

	// Wrap the base handler with Prometheus metrics instrumentation
	metricsWrappedHandler := promhttp.InstrumentHandlerCounter(
		metrics.HTTPRequestsTotal.MustCurryWith(prometheus.Labels{"handler": path}),
		promhttp.InstrumentHandlerDuration(
			metrics.HTTPRequestDuration.MustCurryWith(prometheus.Labels{"handler": path}),
			baseHandler,
		),
	)
	mux.Handle(path, metricsWrappedHandler)
	mux.Handle("/metrics", promhttp.Handler()) // Expose Prometheus metrics
	mux.HandleFunc("/balance", balanceHandler.GetBalance)

	return &App{
		cardCraftAiAPIURL: cfg.CardCraftAi.APIURL,
		httpClient:        &http.Client{},
		cfg:               cfg,
		mux:               mux,
		balanceStorage:    balanceStorage,
	}
}

func (a *App) Run() error {
	addr := ":" + strconv.Itoa(a.cfg.HTTP.Port)
	log.Printf("Starting ConnectRPC server on %s", addr)
	log.Printf("CardCraftAI API URL: %s", a.cardCraftAiAPIURL)

	return http.ListenAndServe(addr, a.mux)
}
