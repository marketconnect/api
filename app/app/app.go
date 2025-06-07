package app

import (
	"api/gen/api/v1/apiv1connect"
	"log"
	"net/http"
	"strconv"

	"api/app/domain/services"
	"api/app/domain/usecases"
	"api/app/internal/config"
	"api/app/internal/presentation"

	"api/app/internal/infrastructure/card_craft_ai"
	"api/app/internal/infrastructure/ozon"
	"api/app/internal/infrastructure/wb"

	"api/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	cardCraftAiAPIURL string
	httpClient        *http.Client
	cfg               *config.Config
	mux               *http.ServeMux
}

// NewApp creates a new ProductServer instance
func NewApp() *App {

	cfg := config.GetConfig()

	// clients
	cardCraftAiClient := card_craft_ai.NewCardCraftAiClient(cfg.CardCraftAi.APIURL)
	wbClient := wb.NewWBClient()
	ozonClient := ozon.NewClient()

	// services
	cardCraftAiService := services.NewCardCraftAiService(cardCraftAiClient)
	wbService := services.NewWbService(cfg.WB.GetCardListMaxAttempts, wbClient)
	ozonService := services.NewOzonService(ozonClient)

	// usecases
	createCardUsecase := usecases.NewCreateCardUsecase(cardCraftAiService, wbService, ozonService)

	// handlers
	createProductCardHandler := presentation.NewCreateProductCardHandler(createCardUsecase)

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

	return &App{
		cardCraftAiAPIURL: cfg.CardCraftAi.APIURL,
		httpClient:        &http.Client{},
		cfg:               cfg,
		mux:               mux,
	}
}

func (a *App) Run() error {
	addr := ":" + strconv.Itoa(a.cfg.HTTP.Port)
	log.Printf("Starting ConnectRPC server on %s", addr)
	log.Printf("CardCraftAI API URL: %s", a.cardCraftAiAPIURL)

	return http.ListenAndServe(addr, a.mux)
}
