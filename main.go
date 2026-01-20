package main

//go:generate go tool swag init

import (
	"log"

	_ "github.com/fachschaftinformatik/web/docs"
	"github.com/fachschaftinformatik/web/internal/api"
)

// @title Fachschaft Informatik API
// @version 1.0
// @description API for the website of the FSV Informatik
// @BasePath /api
func main() {
	if err := api.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
