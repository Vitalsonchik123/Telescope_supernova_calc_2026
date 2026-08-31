package api

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "supernova-calc/internal/app/handler"
    "supernova-calc/internal/app/repository"
)

func StartServer() {
    log.Println("Starting server...")

    repo, err := repository.NewRepository()
    if err != nil {
        logrus.Fatal("Ошибка инициализации репозитория:", err)
    }

    h := handler.NewHandler(repo)

    r := gin.Default()
    r.LoadHTMLGlob("templates/*")
    r.Static("/static", "./resources")

    r.GET("/feed", h.FeedHandler)
    r.GET("/add", h.DraftHandler)
    r.GET("/grid", h.GridHandler)

    r.Run()
    log.Println("Server down")
}
