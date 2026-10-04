package models

import (
	"time"
)

// User – таблица пользователей
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Password  string    `gorm:"type:varchar(100);default:'stub_password'" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Telescope – таблица услуг (телескопов)
type Telescope struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Status      string    `gorm:"default:'draft'" json:"status"`
	ImageKey    string    `json:"image_key"`
	VideoKey    string    `json:"video_key"`
	ApertureCm  int       `gorm:"not null" json:"aperture_cm"`
	FovDeg      float64   `gorm:"not null" json:"fov_deg"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Like – таблица лайков (многие-ко-многим)
type Like struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	TelescopeID uint      `gorm:"not null" json:"telescope_id"`
	User        User      `gorm:"foreignKey:UserID" json:"-"`
	Telescope   Telescope `gorm:"foreignKey:TelescopeID" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}
