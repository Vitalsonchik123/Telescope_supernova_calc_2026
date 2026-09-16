// отвечает за настройку GIN и запуск HTTP сервера

package api

import (
	"log"

	"github.com/gin-gonic/gin"

	"supernova-calc/internal/app/database"
	"supernova-calc/internal/app/handler"
	"supernova-calc/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server...")

	// Инициализация БД (AutoMigrate)
	database.InitDB()

	repo := repository.NewRepository(database.DB)
	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/feed", h.FeedHandler)
	r.GET("/add", h.DraftHandler)
	r.GET("/grid", h.GridHandler)

	r.POST("/create-draft", h.CreateDraftHandler)
	r.POST("/publish", h.PublishHandler)
	r.POST("/delete", h.DeleteHandler)

	r.Run()
	log.Println("Server down")
}
