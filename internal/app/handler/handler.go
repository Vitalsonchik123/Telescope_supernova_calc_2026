package handler

import (
	"net/http"
	"strconv"
	"supernova-calc/internal/app/models"
	"supernova-calc/internal/app/repository"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repo: r}
}

func getMinioURL(key string) string {
	return "http://localhost:9000/supernova-data/" + key
}

// FeedHandler – лента по ID (GET /feed?id=xxx&next=true)
func (h *Handler) FeedHandler(c *gin.Context) {
	id := c.Query("id")
	next := c.Query("next") == "true"
	prev := c.Query("prev") == "true"

	if id == "" {
		c.String(http.StatusBadRequest, "Не указан ID телескопа")
		return
	}

	// получаем текущий телескоп (проверяем, что он существует и не archived)
	_, err := h.Repo.GetByID(id)
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusNotFound, "Телескоп не найден")
		return
	}

	// получаем все опубликованные телескопы (в порядке, который нам нужен)
	all := h.Repo.GetAll()
	if len(all) == 0 {
		c.String(http.StatusNotFound, "Нет доступных телескопов")
		return
	}

	// находим индекс текущего в списке
	currentIndex := -1
	for i, t := range all {
		if t.ID == id {
			currentIndex = i
			break
		}
	}

	// если текущий не найден в списке (например, он archived), то выдаём ошибку
	if currentIndex == -1 {
		c.String(http.StatusNotFound, "Телескоп не найден в списке опубликованных")
		return
	}

	// обработка перехода
	if next {
		nextIndex := (currentIndex + 1) % len(all) // зацикливание
		c.Redirect(http.StatusFound, "/feed?id="+all[nextIndex].ID)
		return
	}
	if prev {
		prevIndex := (currentIndex - 1 + len(all)) % len(all) // зацикливание
		c.Redirect(http.StatusFound, "/feed?id="+all[prevIndex].ID)
		return
	}

	// если ни next, ни prev – просто показываем текущий
	tel, _ := h.Repo.GetByID(id) // уже проверяли, что существует
	view := models.TelescopeView{
		Telescope: tel,
		ImageURL:  getMinioURL(tel.ImageKey),
		VideoURL:  getMinioURL(tel.VideoKey),
		Likes:     len(tel.LikedBy),
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"telescope": view,
		"time":      time.Now().Format("15:04:05"),
	})
}

// DraftHandler – страница добавления (черновик)
func (h *Handler) DraftHandler(c *gin.Context) {
	clear := c.Query("clear") == "true"

	var draft models.Telescope
	var err error
	if clear {
		// Если clear=true, создаём пустой телескоп с пустыми строками
		draft = models.Telescope{
			ID:             "",
			Name:           "",
			Observatory:    "",
			ApertureCm:     0,
			FovDeg:         "",
			Filters:        []string{},
			DepthMag:       "",
			TimeResolution: "",
			Description:    "",
			Status:         "",
			ImageKey:       "",
			VideoKey:       "",
			LikedBy:        []string{},
			CreatedAt:      "",
		}
	} else {
		draft, err = h.Repo.GetDraft()
		if err != nil {
			logrus.Error(err)
			c.String(http.StatusNotFound, "Черновик не найден")
			return
		}
	}

	view := models.TelescopeView{
		Telescope: draft,
		ImageURL:  getMinioURL(draft.ImageKey),
		VideoURL:  getMinioURL(draft.VideoKey),
		Likes:     len(draft.LikedBy),
	}

	c.HTML(http.StatusOK, "add.html", gin.H{
		"telescope": view,
		"clear":     clear, // чтобы знать, что мы в режиме очистки
	})
}

// GridHandler – список с фильтром по диаметру (GET /grid?min_aperture=...)
func (h *Handler) GridHandler(c *gin.Context) {
	minApertureStr := c.Query("min_aperture")
	var minAperture int
	if minApertureStr != "" {
		var err error
		minAperture, err = strconv.Atoi(minApertureStr)
		if err != nil {
			minAperture = 0
		}
	}

	var telescopes []models.Telescope
	if minAperture > 0 {
		telescopes = h.Repo.FilterByApertureMin(minAperture)
	} else {
		telescopes = h.Repo.GetAll()
	}

	var views []models.TelescopeView
	for _, t := range telescopes {
		views = append(views, models.TelescopeView{
			Telescope: t,
			ImageURL:  getMinioURL(t.ImageKey),
			VideoURL:  getMinioURL(t.VideoKey),
			Likes:     len(t.LikedBy),
		})
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"telescopes":  views,
		"minAperture": minAperture,
		"time":        time.Now().Format("15:04:05"),
	})
}
