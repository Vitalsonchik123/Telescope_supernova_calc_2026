package main

import (
	"log"

	"github.com/joho/godotenv"

	"supernova-calc/internal/api"
)

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env not found")
	}

	api.StartServer()
}
