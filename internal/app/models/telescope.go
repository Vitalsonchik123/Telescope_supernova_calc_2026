package models

type Telescope struct {
    ID            string   `json:"id"`
    Name          string   `json:"name"`
    Observatory   string   `json:"observatory"`
    ApertureCm    int      `json:"aperture_cm"`      // для фильтрации
    FovDeg        string   `json:"fov_deg"`          // может быть "47" или "24x96"
    Filters       []string `json:"filters"`
    DepthMag      string   `json:"depth_mag"`        // например, "g ~ 20,8; r ~ 20,6"
    TimeResolution string  `json:"time_resolution"`  // "всё северное небо каждую ночь"
    Description   string   `json:"description"`
    Status        string   `json:"status"`           // "draft", "published", "archived"
    ImageKey      string   `json:"imageKey"`
    VideoKey      string   `json:"videoKey"`
    LikedBy       []string `json:"liked_by"`
    CreatedAt     string   `json:"createdAt"`
}

// TelescopeView – для передачи в шаблон с дополнительными полями
type TelescopeView struct {
    Telescope
    ImageURL   string
    VideoURL   string
    Likes      int
}
