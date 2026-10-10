package models

// UserSerializer — безопасный вывод пользователя (без пароля)
type UserSerializer struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
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

// TelescopeFullSerializer — для ленты в API
type TelescopeFullSerializer struct {
	Telescope
	Creator    UserSerializer `json:"creator"`
	ImageURL   string         `json:"image_url"`
	VideoURL   string         `json:"video_url"`
	LikesCount int            `json:"likes_count"`
	IsMine     bool           `json:"is_mine"`
	IsLiked    bool           `json:"is_liked"`
}

// CreateTelescopeRequest — POST /api/telescopes
type CreateTelescopeRequest struct {
	Name        string  `json:"name" binding:"required"`
	ApertureCm  int     `json:"aperture_cm"`
	FovDeg      float64 `json:"fov_deg"`
	Description string  `json:"description"`
}

// PublishTelescopeRequest — PUT /api/telescopes/:id/publish
type PublishTelescopeRequest struct {
	Description string  `json:"description"`
	ApertureCm  int     `json:"aperture_cm" binding:"required"`
	FovDeg      float64 `json:"fov_deg" binding:"required"`
}

// LikeRequest — POST /api/telescopes/:id/like
type LikeRequest struct {
	Like int `json:"like"`
}

// RegisterRequest — POST /api/users/register
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"` // опционально, по умолчанию "user"
}

// LoginRequest — POST /api/users/login
type LoginRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse — ответ на успешный логин
type LoginResponse struct {
	Token string         `json:"token"`
	User  UserSerializer `json:"user"`
}

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

// TelescopeView — для SSR-шаблонов
type TelescopeView struct {
	Telescope
	ImageURL string
	VideoURL string
	Likes    int
}
