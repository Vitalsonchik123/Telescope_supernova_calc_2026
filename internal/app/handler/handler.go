package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"supernova-calc/internal/app/models"
	redisclient "supernova-calc/internal/app/redis"
	"supernova-calc/internal/app/repository"
)

// Handler — общий контейнер для всех обработчиков
type Handler struct {
	Repo  *repository.Repository
	Redis *redisclient.Client // ← добавлено
}

// NewHandler — конструктор Handler
func NewHandler(r *repository.Repository, redis *redisclient.Client) *Handler {
	return &Handler{
		Repo:  r,
		Redis: redis,
	}
}

// ================================================================
// Вспомогательные функции
// ================================================================

// getMinioURL формирует полный URL к файлу в MinIO (для опубликованных)
func getMinioURL(key string) string {
	return "http://localhost:9000/supernova-data/" + key
}

// getLocalURL формирует локальный URL к файлу-заглушке черновика
func getLocalURL(filename string) string {
	return "/static/default/" + filename
}

// atoiSafe — безопасный Atoi
func atoiSafe(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// parseFloatSafe — безопасный ParseFloat
func parseFloatSafe(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ================================================================
// SSR-обработчики (ЛР2) — рендеринг HTML-шаблонов
// ================================================================

// FeedHandler – страница «Лента»
func (h *Handler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")
	next := c.Query("next") == "true"

	if idStr == "" {
		all, err := h.Repo.GetAll()
		if err != nil || len(all) == 0 {
			c.String(http.StatusNotFound, "Нет доступных телескопов")
			return
		}
		c.Redirect(http.StatusFound, "/feed?id="+strconv.FormatUint(uint64(all[0].ID), 10))
		return
	}

	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}
	id := uint(id64)

	tel, err := h.Repo.GetByID(id)
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusNotFound, "Телескоп не найден")
		return
	}

	all, err := h.Repo.GetAll()
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка БД")
		return
	}
	if len(all) == 0 {
		c.String(http.StatusNotFound, "Нет доступных телескопов")
		return
	}

	currentIndex := -1
	for i, t := range all {
		if t.ID == id {
			currentIndex = i
			break
		}
	}
	if currentIndex == -1 {
		c.Redirect(http.StatusFound, "/feed?id="+strconv.FormatUint(uint64(all[0].ID), 10))
		return
	}

	if next {
		ni := (currentIndex + 1) % len(all)
		c.Redirect(http.StatusFound, "/feed?id="+strconv.FormatUint(uint64(all[ni].ID), 10))
		return
	}

	likes, _ := h.Repo.GetLikesCount(tel.ID)
	view := struct {
		ID          uint
		Name        string
		ApertureCm  int
		FovDeg      float64
		Description string
		ImageURL    string
		VideoURL    string
		Likes       int
	}{
		ID:          tel.ID,
		Name:        tel.Name,
		ApertureCm:  tel.ApertureCm,
		FovDeg:      tel.FovDeg,
		Description: tel.Description,
		ImageURL:    getMinioURL(tel.ImageKey),
		VideoURL:    getMinioURL(tel.VideoKey),
		Likes:       int(likes),
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"telescope": view,
		"time":      time.Now().Format("15:04:05"),
	})
}

// DraftHandler – страница «Добавить»
func (h *Handler) DraftHandler(c *gin.Context) {
	const currentUserID uint = 1
	clear := c.Query("clear") == "true"

	var draft *struct {
		ID          uint
		Name        string
		Description string
		ApertureCm  int
		FovDeg      float64
		ImageKey    string
		VideoKey    string
		ImageURL    string
		VideoURL    string
	}

	if clear {
		draft = &struct {
			ID          uint
			Name        string
			Description string
			ApertureCm  int
			FovDeg      float64
			ImageKey    string
			VideoKey    string
			ImageURL    string
			VideoURL    string
		}{}
	} else {
		d, err := h.Repo.GetDraft(currentUserID)
		if err != nil {
			logrus.Error(err)
			c.String(http.StatusInternalServerError, "Ошибка БД")
			return
		}
		if d != nil {
			draft = &struct {
				ID          uint
				Name        string
				Description string
				ApertureCm  int
				FovDeg      float64
				ImageKey    string
				VideoKey    string
				ImageURL    string
				VideoURL    string
			}{
				ID:          d.ID,
				Name:        d.Name,
				Description: d.Description,
				ApertureCm:  d.ApertureCm,
				FovDeg:      d.FovDeg,
				ImageKey:    d.ImageKey,
				VideoKey:    d.VideoKey,
				ImageURL:    getLocalURL("tess_image.jpg"),
				VideoURL:    getLocalURL("tess_video.mp4"),
			}
		} else {
			draft = &struct {
				ID          uint
				Name        string
				Description string
				ApertureCm  int
				FovDeg      float64
				ImageKey    string
				VideoKey    string
				ImageURL    string
				VideoURL    string
			}{
				ImageURL: getLocalURL("tess_image.jpg"),
				VideoURL: getLocalURL("tess_video.mp4"),
			}
		}
	}

	c.HTML(http.StatusOK, "add.html", gin.H{
		"telescope": draft,
		"hasDraft":  draft.ID != 0,
	})
}

// GridHandler – страница «Плитка»
func (h *Handler) GridHandler(c *gin.Context) {
	minStr := c.Query("min_aperture")
	minAperture := 0
	applied := false

	if minStr != "" {
		if v, err := strconv.Atoi(minStr); err == nil {
			minAperture = v
			applied = true
		}
	}

	var telescopes []models.Telescope
	var err error
	if applied && minAperture > 0 {
		telescopes, err = h.Repo.FilterByApertureMin(minAperture)
	} else {
		telescopes, err = h.Repo.GetAll()
	}
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка БД")
		return
	}

	var views []models.TelescopeView
	for _, t := range telescopes {
		likes, _ := h.Repo.GetLikesCount(t.ID)
		views = append(views, models.TelescopeView{
			Telescope: t,
			ImageURL:  getMinioURL(t.ImageKey),
			VideoURL:  getMinioURL(t.VideoKey),
			Likes:     int(likes),
		})
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"telescopes":  views,
		"minAperture": minAperture,
		"applied":     applied,
		"time":        time.Now().Format("15:04:05"),
	})
}

// CreateDraftHandler – POST /create-draft (SSR)
func (h *Handler) CreateDraftHandler(c *gin.Context) {
	const currentUserID uint = 1

	existing, err := h.Repo.GetDraft(currentUserID)
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка БД")
		return
	}
	if existing != nil {
		c.Redirect(http.StatusFound, "/add")
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "Не указано название")
		return
	}

	t := &models.Telescope{
		Name:   name,
		UserID: currentUserID,
		Status: "draft",
	}

	if err := h.Repo.CreateDraft(t); err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка создания черновика")
		return
	}

	c.Redirect(http.StatusFound, "/add")
}

// PublishHandler – POST /publish (SSR)
func (h *Handler) PublishHandler(c *gin.Context) {
	idStr := c.PostForm("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}
	id := uint(id64)

	updates := map[string]interface{}{
		"description": c.PostForm("description"),
		"aperture_cm": atoiSafe(c.PostForm("aperture")),
		"fov_deg":     parseFloatSafe(c.PostForm("fov")),
	}

	if err := h.Repo.UpdateFields(id, updates); err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка обновления")
		return
	}

	if err := h.Repo.Publish(id); err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка публикации")
		return
	}

	c.Redirect(http.StatusFound, "/grid")
}

// DeleteHandler – POST /delete (SSR)
func (h *Handler) DeleteHandler(c *gin.Context) {
	idStr := c.PostForm("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}

	if err := h.Repo.Delete(uint(id64)); err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка удаления")
		return
	}

	c.Redirect(http.StatusFound, "/grid")
}
