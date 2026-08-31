package repository

import (
    "errors"
    "strconv"
    "strings"
    "supernova-calc/internal/app/models"
)

type Repository struct {
    telescopes []models.Telescope
}

func NewRepository() (*Repository, error) {
    telescopes := []models.Telescope{
        {
            ID:            "ztf-001",
            Name:          "ZTF",
            Observatory:   "Паломарская обсерватория, Калифорния, США",
            ApertureCm:    122,
            FovDeg:        "47",
            Filters:       []string{"g", "r", "i"},
            DepthMag:      "g ~ 20,8; r ~ 20,6 (5σ, 30 с)",
            TimeResolution: "всё северное небо каждую ночь",
            Description:   "Широкоугольный обзорный телескоп для поиска транзиентных событий.",
            Status:        "published",
            ImageKey:      "ztf_image.jpg",
            VideoKey:      "ztf_video.mp4",
            LikedBy:       []string{"user1", "user3", "user7"},
            CreatedAt:     "2025-01-01",
        },
        // Pan-STARRS
        {
            ID:            "panstarrs-001",
            Name:          "Pan-STARRS",
            Observatory:   "Обсерватория Халеакала, Мауи, Гавайи, США",
            ApertureCm:    180,
            FovDeg:        "7",
            Filters:       []string{"g", "r", "i", "z", "y"},
            DepthMag:      "r ~ 22 (одиночный снимок)",
            TimeResolution: "~500–600 снимков/ночь",
            Description:   "Панорамный обзорный телескоп для поиска астероидов и транзиентов.",
            Status:        "published",
            ImageKey:      "panstarrs_image.jpg",
            VideoKey:      "panstarrs_video.mp4",
            LikedBy:       []string{"user2", "user5"},
            CreatedAt:     "2025-01-02",
        },
        // LSST
        {
            ID:            "lsst-001",
            Name:          "LSST (Обсерватория Веры Рубин)",
            Observatory:   "Серро-Пачон, Чили",
            ApertureCm:    840,
            FovDeg:        "10",
            Filters:       []string{"u", "g", "r", "i", "z", "y"},
            DepthMag:      "до 27-й зв. вел. (за 30 с)",
            TimeResolution: "всё доступное небо каждые 3 ночи",
            Description:   "Будущий мощный обзорный телескоп.",
            Status:        "published",
            ImageKey:      "lsst_image.jpg",
            VideoKey:      "lsst_video.mp4",
            LikedBy:       []string{"user1", "user4", "user6", "user8"},
            CreatedAt:     "2025-01-03",
        },
        // TESS – черновик
        {
            ID:            "tess-001",
            Name:          "TESS",
            Observatory:   "Орбита Земли (низкая орбита)",
            ApertureCm:    10, // 10.5 см, округлим до 10 для целого числа
            FovDeg:        "24x96 (4 камеры)",
            Filters:       []string{"600–1000 нм (красный/ИК)"},
            DepthMag:      "до 15-й зв. вел.",
            TimeResolution: "27 дней на сектор",
            Description:   "Космический телескоп для поиска экзопланет, также обнаруживает сверхновые.",
            Status:        "draft",
            ImageKey:      "tess_image.jpg",
            VideoKey:      "tess_video.mp4",
            LikedBy:       []string{"user2", "user3"},
            CreatedAt:     "2025-01-04",
        },
        // Gaia
        {
            ID:            "gaia-001",
            Name:          "Gaia",
            Observatory:   "Точка Лагранжа L2 системы Земля-Солнце",
            ApertureCm:    145, // 145×50 см, берём 145 как основное число
            FovDeg:        "0,5",
            Filters:       []string{"G", "BP", "RP (фотометрия)"},
            DepthMag:      "до 20-й зв. вел.",
            TimeResolution: "непрерывный обзор",
            Description:   "Астрометрический телескоп, данные используются для кривых блеска.",
            Status:        "published",
            ImageKey:      "gaia_image.jpg",
            VideoKey:      "gaia_video.mp4",
            LikedBy:       []string{"user5", "user7", "user9"},
            CreatedAt:     "2025-01-05",
        },
        // Keck II – архивирован
        {
            ID:            "keck-001",
            Name:          "Keck II",
            Observatory:   "Обсерватория Мауна-Кеа, Гавайи, США",
            ApertureCm:    1000,
            FovDeg:        "Узкое поле (спектроскопия)",
            Filters:       []string{"DEIMOS", "MOSFIRE", "NIRSPEC"},
            DepthMag:      "спектроскопия до 24-й зв. вел.",
            TimeResolution: "по запросу",
            Description:   "Один из крупнейших оптических телескопов для спектроскопии.",
            Status:        "archived",
            ImageKey:      "keck_image.jpg",
            VideoKey:      "keck_video.mp4",
            LikedBy:       []string{"user1", "user6"},
            CreatedAt:     "2025-01-06",
        },
    }
    return &Repository{telescopes: telescopes}, nil
}

// Методы репозитория (GetAll, GetByID, GetDraft, FilterByApertureMin) – без изменений
// (убедитесь, что они используют поле ApertureCm для фильтрации)
