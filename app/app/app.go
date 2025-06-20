package app

import (
	"api/gen/api/v1/apiv1connect"
	"context"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"api/app/domain/services"
	"api/app/domain/usecases"
	"api/app/internal/config"
	"api/app/internal/presentation"
	"api/app/internal/presentation/middleware"

	"api/app/internal/infrastructure/external/card_craft_ai"
	"api/app/internal/infrastructure/external/ozon"
	"api/app/internal/infrastructure/external/token_counter"
	"api/app/internal/infrastructure/external/wb"
	"api/app/internal/infrastructure/file_storage"
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
	cardCraftAiClient := card_craft_ai.NewCardCraftAiClient("http://" + cfg.CardCraftAi.URL + ":" + strconv.Itoa(cfg.CardCraftAi.Port))
	wbClient := wb.NewWBClient()
	ozonClient := ozon.NewClient()
	tokenCounterClient := token_counter.NewClient("http://" + cfg.TokenCounter.APIURL + ":" + strconv.Itoa(cfg.TokenCounter.Port))

	// file storage client - configure upload directory and base URL
	uploadDir := cfg.FileStorage.UploadDir // Directory to store uploaded files

	// Generate the correct URL path from upload directory
	var urlPath string
	if strings.HasPrefix(uploadDir, "./") {
		// Remove "./" prefix for URL path: "./uploads" -> "/uploads"
		urlPath = "/" + strings.TrimPrefix(uploadDir, "./")
	} else if strings.HasPrefix(uploadDir, "/") {
		// Absolute path - use last component for URL: "/var/uploads" -> "/uploads"
		urlPath = "/" + filepath.Base(uploadDir)
	} else {
		// Relative path - use as URL path: "uploads" -> "/uploads"
		urlPath = "/" + uploadDir
	}

	// Determine base URL - use configured public URL or fallback to localhost
	var baseURL string
	if cfg.FileStorage.BaseURL != "" {
		// Use configured public URL (for production): "https://myapi.com" + "/uploads"
		baseURL = strings.TrimSuffix(cfg.FileStorage.BaseURL, "/") + urlPath
		log.Printf("Using configured public baseURL for file storage: %s", baseURL)
	} else {
		// Fallback to localhost (for development)
		baseURL = "http://localhost:" + strconv.Itoa(cfg.HTTP.Port) + urlPath
		log.Printf("WARNING: Using localhost baseURL for file storage: %s - Set FILE_STORAGE_BASE_URL for production!", baseURL)
	}

	fileTTL := time.Duration(cfg.FileStorage.TTLMinutes) * time.Minute // Keep files for configured hours
	fileStorageClient := file_storage.NewTemporaryFileStorage(uploadDir, baseURL, fileTTL)

	// services
	cardCraftAiService := services.NewCardCraftAiService(cardCraftAiClient)
	tokenBillingService := services.NewTokenBillingService(tokenCounterClient, balanceStorage)
	wbService := services.NewWbService(cfg.WB.GetCardListMaxAttempts, wbClient)
	fileUploadService := services.NewFileUploadService(fileStorageClient)
	ozonService := services.NewOzonService(ozonClient, fileUploadService)

	// usecases
	createCardUsecase := usecases.NewCreateCardUsecase(cardCraftAiService, wbService, ozonService, tokenBillingService)
	getBalanceUsecase := usecases.NewGetBalanceUsecase(balanceStorage)
	updateBalanceUsecase := usecases.NewUpdateBalanceUsecase(balanceStorage)

	// handlers
	createProductCardHandler := presentation.NewCreateProductCardHandler(createCardUsecase)
	balanceHandler := presentation.NewBalanceHandler(getBalanceUsecase)
	tinkoffHandler := presentation.NewTinkoffNotificationHandler(
		updateBalanceUsecase,
		cfg.Tinkoff.SecretKey,
		cfg.Tinkoff.TerminalKey,
		cfg.Tinkoff.TelegramBotToken,
	)

	// middleware
	balanceCheckMiddleware := middleware.NewBalanceCheckMiddleware(getBalanceUsecase, presentation.ExtractAPIKeyFromHeader)

	// mux
	mux := http.NewServeMux()
	path, baseHandler := apiv1connect.NewProductServiceHandler(createProductCardHandler)
	balancePath, balanceServiceHandler := apiv1connect.NewBalanceServiceHandler(balanceHandler)
	paymentPath, paymentServiceHandler := apiv1connect.NewPaymentServiceHandler(tinkoffHandler)

	// Wrap the base handler with balance check and Prometheus metrics instrumentation
	balanceCheckedHandler := balanceCheckMiddleware.CheckBalance(baseHandler)
	metricsWrappedHandler := promhttp.InstrumentHandlerCounter(
		metrics.HTTPRequestsTotal.MustCurryWith(prometheus.Labels{"handler": path}),
		promhttp.InstrumentHandlerDuration(
			metrics.HTTPRequestDuration.MustCurryWith(prometheus.Labels{"handler": path}),
			balanceCheckedHandler,
		),
	)

	balanceMetricsWrappedHandler := promhttp.InstrumentHandlerCounter(
		metrics.HTTPRequestsTotal.MustCurryWith(prometheus.Labels{"handler": balancePath}),
		promhttp.InstrumentHandlerDuration(
			metrics.HTTPRequestDuration.MustCurryWith(prometheus.Labels{"handler": balancePath}),
			balanceServiceHandler,
		),
	)

	mux.Handle(path, metricsWrappedHandler)
	mux.Handle(balancePath, balanceMetricsWrappedHandler)
	mux.Handle(paymentPath, paymentServiceHandler)
	mux.Handle("/metrics", promhttp.Handler()) // Expose Prometheus metrics
	mux.HandleFunc("/balance", balanceHandler.GetBalanceHTTP)
	mux.HandleFunc("/balance-by-token", balanceHandler.GetBalanceByToken)

	// Tinkoff payment endpoints
	mux.HandleFunc("/payment/request", tinkoffHandler.ProcessPaymentRequestHandler)
	mux.HandleFunc("/payment/notification", tinkoffHandler.TinkoffNotificationHandler)

	// Serve uploaded files statically - use the same URL path as generated above
	mux.Handle(urlPath+"/", http.StripPrefix(urlPath+"/", http.FileServer(http.Dir(uploadDir))))

	return &App{
		cardCraftAiAPIURL: cfg.CardCraftAi.URL,
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
