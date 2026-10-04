package models

// ================================================================
// Сериализаторы (DTO) — для передачи данных по API
// ================================================================

// UserSerializer — безопасный вывод пользователя (без пароля)
type UserSerializer struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// TelescopeListSerializer — краткая информация для списка услуг
type TelescopeListSerializer struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Observatory string  `json:"observatory"`
	ApertureCm  int     `json:"aperture_cm"`
	FovDeg      float64 `json:"fov_deg"`
	ImageURL    string  `json:"image_url"`
	LikesCount  int     `json:"likes_count"`
	IsMine      bool    `json:"is_mine"`  // создал ли текущий пользователь эту услугу
	IsLiked     bool    `json:"is_liked"` // поставил ли текущий пользователь лайк
}

// TelescopeFullSerializer — подробная информация (с creator)
type TelescopeFullSerializer struct {
	Telescope
	Creator    UserSerializer `json:"creator"`
	ImageURL   string         `json:"image_url"`
	VideoURL   string         `json:"video_url"`
	LikesCount int            `json:"likes_count"`
	IsMine     bool           `json:"is_mine"`  // создал ли текущий пользователь
	IsLiked    bool           `json:"is_liked"` // поставил ли текущий пользователь лайк
}

// ================================================================
// DTO для входящих запросов
// ================================================================

// CreateTelescopeRequest — что приходит от клиента при POST /api/telescopes
type CreateTelescopeRequest struct {
	Name        string  `json:"name" binding:"required"`
	Observatory string  `json:"observatory"`
	ApertureCm  int     `json:"aperture_cm"`
	FovDeg      float64 `json:"fov_deg"`
	Description string  `json:"description"`
}

// PublishTelescopeRequest — что приходит при PUT /api/telescopes/:id/publish
type PublishTelescopeRequest struct {
	Description string  `json:"description"`
	Observatory string  `json:"observatory"`
	ApertureCm  int     `json:"aperture_cm"`
	FovDeg      float64 `json:"fov_deg"`
}

// LikeRequest — что приходит при POST /api/telescopes/:id/like
type LikeRequest struct {
	Like int `json:"like"` // 1 — поставить, 0 — отменить
}

// RegisterRequest — POST /api/users/register
type RegisterRequest struct {
	Name string `json:"name" binding:"required"`
}

// LoginRequest — POST /api/users/login (заглушка)
type LoginRequest struct {
	Name string `json:"name" binding:"required"`
}

// ================================================================
// Форматы ответов
// ================================================================

// APIError — формат ответа об ошибке
type APIError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// APISuccess — формат успешного ответа
type APISuccess struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
