package main

import (
	"log"

	"github.com/joho/godotenv"

	"supernova-calc/internal/api"
)

// @title Supernova Calc API
// @version 1.0
// @description REST API для системы расчёта энергии вспышек сверхновых типа Ia
// @contact.name API Support
// @contact.email support@supernova.local
// @license.name AS IS (NO WARRANTY)
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env not found")
	}

	api.StartServer()
}
