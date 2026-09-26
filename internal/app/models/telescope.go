package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Telescope struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Observatory string    `json:"observatory"`
	ApertureCm  int       `json:"aperture_cm"`
	FovDeg      float64   `json:"fov_deg"`
	Description string    `json:"description"`
	Status      string    `gorm:"default:'draft'" json:"status"`
	ImageKey    string    `json:"image_key"`
	VideoKey    string    `json:"video_key"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Like struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	TelescopeID uint      `gorm:"not null" json:"telescope_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
	Telescope   Telescope `gorm:"foreignKey:TelescopeID" json:"telescope"`
	CreatedAt   time.Time `json:"created_at"`
}

type TelescopeView struct {
	Telescope
	ImageURL string
	VideoURL string
	Likes    int
}
