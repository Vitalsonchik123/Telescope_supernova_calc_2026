package api

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"supernova-calc/internal/app/auth"
	"supernova-calc/internal/app/database"
	"supernova-calc/internal/app/handler"
	"supernova-calc/internal/app/repository"

	_ "supernova-calc/docs" // ← генерируется swag init
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
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// ========== Swagger UI ==========
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ========== SSR-маршруты (ЛР2) ==========
	r.GET("/feed", h.FeedHandler)
	r.GET("/add", h.DraftHandler)
	r.GET("/grid", h.GridHandler)
	r.POST("/create-draft", h.CreateDraftHandler)
	r.POST("/publish", h.PublishHandler)
	r.POST("/delete", h.DeleteHandler)

	// ========== API-маршруты (ЛР3) ==========
	api := r.Group("/api")
	{
		// Домен услуг
		api.GET("/telescopes", h.APIGetTelescopes)
		api.GET("/feed", h.APIFeed)
		api.GET("/draft", h.APIGetDraft)
		api.POST("/telescopes", h.APICreateTelescope)
		api.PUT("/telescopes/:id/publish", h.APIPublishTelescope)
		api.DELETE("/telescopes/:id", h.APIDeleteTelescope)
		api.POST("/telescopes/:id/like", h.APILikeTelescope)

		// Домен пользователей
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
