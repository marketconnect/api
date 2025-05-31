package main

import (
	"api/app/app"
	"log"
)

func main() {
	a := app.NewApp()

	if err := a.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
