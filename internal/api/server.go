package api

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"supernova-calc/internal/app/auth"
	"supernova-calc/internal/app/database"
	"supernova-calc/internal/app/handler"
	redisclient "supernova-calc/internal/app/redis"
	"supernova-calc/internal/app/repository"

	_ "supernova-calc/docs"
)

// RedisClient — глобальный клиент Redis (пригодится для middleware)
var RedisClient *redisclient.Client

func StartServer() {
	log.Println("Starting server...")

	// 1. Подключение к БД
	database.InitDB()

	// 2. Подключение к Redis
	ctx := context.Background()
	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	if os.Getenv("REDIS_HOST") == "" {
		redisAddr = "localhost:6379"
	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	if redisPass == "" {
		redisPass = "password"
	}

	rdb, err := redisclient.NewClient(ctx, redisAddr, redisPass)
	if err != nil {
		logrus.Fatal("Redis init failed: ", err)
	}
	RedisClient = rdb
	log.Println("Redis connected")

	// 3. Репозиторий
	repo := repository.NewRepository(database.DB)

	// 4. MinIO
	err = repo.InitMinio(
		os.Getenv("MINIO_ENDPOINT"),
		os.Getenv("MINIO_ACCESS_KEY"),
		os.Getenv("MINIO_SECRET_KEY"),
		os.Getenv("MINIO_BUCKET"),
	)
	if err != nil {
		logrus.Fatal("Minio init failed: ", err)
	}

	// 5. Singleton-пользователь (пока оставляем)
	auth.InitCurrentUser(repo)

	// 6. Handler — теперь с Redis
	h := handler.NewHandler(repo, rdb)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// Swagger UI
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
		// Публичные — доступны без токена
		api.GET("/telescopes", h.APIGetTelescopes)
		api.GET("/feed", h.APIFeed)
		api.POST("/users/register", h.APIRegister)
		api.POST("/users/login", h.APILogin)

		// Защищённые — требуют JWT (пока без middleware, чтобы не ломать ЛР3)
		api.GET("/draft", h.APIGetDraft)
		api.POST("/telescopes", h.APICreateTelescope)
		api.PUT("/telescopes/:id/publish", h.APIPublishTelescope)
		api.DELETE("/telescopes/:id", h.APIDeleteTelescope)
		api.POST("/telescopes/:id/like", h.APILikeTelescope)
		api.POST("/users/logout", h.APILogout)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
	log.Println("Server down")
}
