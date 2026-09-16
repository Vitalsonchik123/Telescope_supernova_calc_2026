//подключение к PostgreSQL и автоматические миграции

package database

import (
	"log"
	"supernova-calc/internal/app/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost user=postgres password=postgres dbname=supernova_db port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// автоматическая миграция (создание таблиц)
	err = DB.AutoMigrate(&models.User{}, &models.Telescope{}, &models.Like{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Database connected and migrated")
}
