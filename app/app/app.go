package app

import (
	"api/gen/api/v1/apiv1connect"
	"log"
	"net/http"
	"strconv"

	"api/app/domain/usecases"
	"api/app/internal/config"
	"api/app/internal/infrastructure/card_craft_ai"
	"api/app/internal/infrastructure/wb"
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

	// usecases
	createCardUsecase := usecases.NewCreateCardUsecase(cardCraftAiClient, wbClient, cfg.WB.GetCardListMaxAttempts)

	// mux
	mux := http.NewServeMux()
	path, handler := apiv1connect.NewCreateProductCardServiceHandler(createCardUsecase)
	mux.Handle(path, handler)

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
