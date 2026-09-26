// отвечает за настройку GIN и запуск HTTP сервера

package api

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"supernova-calc/internal/app/auth"
	"supernova-calc/internal/app/database"
	"supernova-calc/internal/app/handler"
	"supernova-calc/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server...")

	// Подключение к БД + AutoMigrate
	database.InitDB()

	// Репозиторий
	repo := repository.NewRepository(database.DB)

	// MinIO
	err := repo.InitMinio(
		os.Getenv("MINIO_ENDPOINT"),
		os.Getenv("MINIO_ACCESS_KEY"),
		os.Getenv("MINIO_SECRET_KEY"),
		os.Getenv("MINIO_BUCKET"),
	)
	if err != nil {
		logrus.Fatal("Minio init failed: ", err)
	}

	// Singleton-пользователь
	auth.InitCurrentUser(repo)

	// Handler
	h := handler.NewHandler(repo)

	r := gin.Default()

	// ========== SSR (ЛР2) ==========
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/feed", h.FeedHandler)
	r.GET("/add", h.DraftHandler)
	r.GET("/grid", h.GridHandler)
	r.POST("/create-draft", h.CreateDraftHandler)
	r.POST("/publish", h.PublishHandler)
	r.POST("/delete", h.DeleteHandler)

	// ========== API (ЛР3) ==========
	api := r.Group("/api")
	{
		// Услуги
		api.GET("/telescopes", h.APIGetTelescopes)                // список с фильтром
		api.GET("/feed", h.APIFeed)                               // лента
		api.GET("/draft", h.APIGetDraft)                          // черновик
		api.POST("/telescopes", h.APICreateTelescope)             // создать черновик + файлы
		api.PUT("/telescopes/:id/publish", h.APIPublishTelescope) // публикация
		api.DELETE("/telescopes/:id", h.APIDeleteTelescope)       // soft delete
		api.POST("/telescopes/:id/like", h.APILikeTelescope)      // лайк

		// Пользователи
		api.POST("/users/register", h.APIRegister)
		api.POST("/users/login", h.APILogin)
		api.POST("/users/logout", h.APILogout)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
	log.Println("Server down")
}
