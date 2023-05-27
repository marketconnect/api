package main

import (
	"context"
	"log"
	"mc_api/internal/app"
	"mc_api/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.GetConfig()
	a, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = a.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
