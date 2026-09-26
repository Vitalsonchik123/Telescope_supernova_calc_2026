package main

import (
	"log"

	"github.com/joho/godotenv"

	"supernova-calc/internal/app/database"
	"supernova-calc/internal/app/models"
)

func main() {
	_ = godotenv.Load()
	database.InitDB()

	// AutoMigrate создаёт таблицы
	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Telescope{},
		&models.Like{},
	); err != nil {
		log.Fatal("Migration failed:", err)
	}

	var count int64
	database.DB.Model(&models.User{}).Where("id = ?", 1).Count(&count)
	if count == 0 {
		database.DB.Create(&models.User{Name: "test_user"})
		log.Println("Created default user with id=1")
	}

	log.Println("Migration completed")
}
