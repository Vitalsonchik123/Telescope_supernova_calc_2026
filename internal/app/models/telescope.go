//описание структуры таблиц БД через GORM-модели.

package models

import (
	"time"
)

// User – таблица пользователей
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Telescope – таблица услуг (телескопов)
type Telescope struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"not null" json:"name"`
	Observatory    string    `json:"observatory"`
	ApertureCm     int       `json:"aperture_cm"`              // для фильтрации
	FovDeg         string    `json:"fov_deg"`                  // "47" или "24x96"
	Filters        string    `gorm:"type:text" json:"filters"` // храним как строку, например "g,r,i"
	DepthMag       string    `json:"depth_mag"`                // "g ~ 20,8; r ~ 20,6"
	TimeResolution string    `json:"time_resolution"`          // "всё северное небо каждую ночь"
	Description    string    `json:"description"`
	Status         string    `gorm:"default:'draft'" json:"status"` // draft, published, deleted
	ImageKey       string    `json:"image_key"`
	VideoKey       string    `json:"video_key"`
	UserID         uint      `gorm:"not null" json:"user_id"` // создатель
	User           User      `gorm:"foreignKey:UserID" json:"user"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Like – таблица лайков (многие-ко-многим)
type Like struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	TelescopeID uint      `gorm:"not null" json:"telescope_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
	Telescope   Telescope `gorm:"foreignKey:TelescopeID" json:"telescope"`
	CreatedAt   time.Time `json:"created_at"`
}

// TelescopeView – для передачи в шаблон с дополнительными полями
type TelescopeView struct {
	Telescope
	ImageURL     string   // полный URL к изображению
	VideoURL     string   // полный URL к видео
	Likes        int      // количество лайков
	FiltersSlice []string // для удобства отображения в шаблоне
}
