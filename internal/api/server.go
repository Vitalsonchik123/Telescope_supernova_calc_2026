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

var RedisClient *redisclient.Client

func StartServer() {
	log.Println("Starting server...")

	database.InitDB()

	// Redis
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

	// Репозиторий
	repo := repository.NewRepository(database.DB)

	// MinIO
	err = repo.InitMinio(
		os.Getenv("MINIO_ENDPOINT"),
		os.Getenv("MINIO_ACCESS_KEY"),
		os.Getenv("MINIO_SECRET_KEY"),
		os.Getenv("MINIO_BUCKET"),
	)
	if err != nil {
		logrus.Fatal("Minio init failed: ", err)
	}

	// Singleton — оставляем для SSR-части (handler.go)
	auth.InitCurrentUser(repo)

	// Handler
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

	// ========== API-маршруты (ЛР3 + ЛР4) ==========
	api := r.Group("/api")
	{
		// ---------- Публичные (без токена) ----------
		api.GET("/telescopes", h.APIGetTelescopes)
		api.GET("/feed", h.APIFeed)
		api.POST("/users/register", h.APIRegister)
		api.POST("/users/login", h.APILogin)

		// ---------- Защищённые (требуют JWT) ----------
		protected := api.Group("")
		protected.Use(auth.RequireAuth(rdb))
		{
			protected.GET("/draft", h.APIGetDraft)
			protected.POST("/telescopes", h.APICreateTelescope)
			protected.PUT("/telescopes/:id/publish", h.APIPublishTelescope)
			protected.DELETE("/telescopes/:id", h.APIDeleteTelescope)
			protected.POST("/telescopes/:id/like", h.APILikeTelescope)
			protected.POST("/users/logout", h.APILogout)
		}
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
	log.Println("Server down")
}
