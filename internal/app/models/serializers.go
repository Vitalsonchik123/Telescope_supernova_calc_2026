package models

// UserSerializer
type UserSerializer struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// TelescopeListSerializer — для списка услуг в API
type TelescopeListSerializer struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	ApertureCm int     `json:"aperture_cm"`
	FovDeg     float64 `json:"fov_deg"`
	ImageURL   string  `json:"image_url"`
	LikesCount int     `json:"likes_count"`
	IsMine     bool    `json:"is_mine"`
	IsLiked    bool    `json:"is_liked"`
}

// TelescopeFullSerializer — для ленты в API (наследует Telescope → даты в JSON)
type TelescopeFullSerializer struct {
	Telescope
	Creator    UserSerializer `json:"creator"`
	ImageURL   string         `json:"image_url"`
	VideoURL   string         `json:"video_url"`
	LikesCount int            `json:"likes_count"`
	IsMine     bool           `json:"is_mine"`
	IsLiked    bool           `json:"is_liked"`
}

// DTO
type CreateTelescopeRequest struct {
	Name        string  `json:"name" binding:"required"`
	ApertureCm  int     `json:"aperture_cm"`
	FovDeg      float64 `json:"fov_deg"`
	Description string  `json:"description"`
}

type PublishTelescopeRequest struct {
	Description string  `json:"description"`
	ApertureCm  int     `json:"aperture_cm" binding:"required"`
	FovDeg      float64 `json:"fov_deg" binding:"required"`
}

type LikeRequest struct {
	Like int `json:"like"`
}

type RegisterRequest struct {
	Name string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Name string `json:"name" binding:"required"`
}

// Форматы ответов API
type APIError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type APISuccess struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// TelescopeView — для SSR-шаблонов
type TelescopeView struct {
	Telescope
	ImageURL string
	VideoURL string
	Likes    int
}
