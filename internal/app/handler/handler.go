package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"supernova-calc/internal/app/models"
	"supernova-calc/internal/app/repository"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repo: r}
}

// getMinioURL формирует полный URL к файлу в MinIO
func getMinioURL(key string) string {
	return "http://localhost:9000/supernova-data/" + key
}

// splitFilters превращает строку "g,r,i" в срез строк
func splitFilters(f string) []string {
	if f == "" {
		return []string{}
	}
	parts := strings.Split(f, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// buildView собирает TelescopeView из модели Telescope + данные лайков и MinIO
func (h *Handler) buildView(t models.Telescope) models.TelescopeView {
	likes, _ := h.Repo.GetLikesCount(t.ID)
	return models.TelescopeView{
		Telescope:    t,
		ImageURL:     getMinioURL(t.ImageKey),
		VideoURL:     getMinioURL(t.VideoKey),
		Likes:        int(likes),
		FiltersSlice: splitFilters(t.Filters),
	}
}

// -------------------------------------------------------
// GET-обработчики
// -------------------------------------------------------

// FeedHandler – лента по ID
func (h *Handler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")
	next := c.Query("next") == "true"

	// Если ID не передан — берём первый опубликованный телескоп
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

	// ... остальная логика без изменений
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

	view := h.buildView(*tel)
	c.HTML(http.StatusOK, "feed.html", gin.H{
		"telescope": view,
		"time":      time.Now().Format("15:04:05"),
	})
}

// DraftHandler – страница добавления /add
func (h *Handler) DraftHandler(c *gin.Context) {
	// В ЛР2 пока используем фиксированного пользователя id=1
	const currentUserID uint = 1

	clear := c.Query("clear") == "true"

	var view models.TelescopeView

	if clear {
		// Пустая модель для очистки полей
		view = models.TelescopeView{
			Telescope:    models.Telescope{},
			FiltersSlice: []string{},
		}
	} else {
		draft, err := h.Repo.GetDraft(currentUserID)
		if err != nil {
			logrus.Error(err)
			c.String(http.StatusInternalServerError, "Ошибка БД")
			return
		}
		if draft == nil {
			// Черновика нет — отдаём пустую форму (для кнопки «Далее»)
			view = models.TelescopeView{
				Telescope:    models.Telescope{UserID: currentUserID},
				FiltersSlice: []string{},
			}
		} else {
			view = h.buildView(*draft)
		}
	}

	c.HTML(http.StatusOK, "add.html", gin.H{
		"telescope": view,
		"hasDraft":  view.Telescope.ID != 0,
	})
}

// GridHandler – плитка /grid?min_aperture=...
func (h *Handler) GridHandler(c *gin.Context) {
	minStr := c.Query("min_aperture")
	minAperture := 0
	if minStr != "" {
		if v, err := strconv.Atoi(minStr); err == nil {
			minAperture = v
		}
	}

	var telescopes []models.Telescope
	var err error
	if minAperture > 0 {
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
		views = append(views, h.buildView(t))
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"telescopes":  views,
		"minAperture": minAperture,
		"time":        time.Now().Format("15:04:05"),
	})
}

// -------------------------------------------------------
// POST-обработчики (ЛР2)
// -------------------------------------------------------

// CreateDraftHandler – создание черновика (POST /create-draft)
// Принимает: name, image_key, video_key
func (h *Handler) CreateDraftHandler(c *gin.Context) {
	const currentUserID uint = 1

	name := c.PostForm("name")
	imageKey := c.PostForm("image_key")
	videoKey := c.PostForm("video_key")

	if name == "" {
		c.String(http.StatusBadRequest, "Не указано название")
		return
	}

	t := &models.Telescope{
		Name:     name,
		ImageKey: imageKey,
		VideoKey: videoKey,
		UserID:   currentUserID,
		Status:   "draft",
	}

	if err := h.Repo.CreateDraft(t); err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка создания черновика")
		return
	}

	c.Redirect(http.StatusFound, "/add")
}

// PublishHandler – публикация услуги (POST /publish)
// Принимает: id, description, aperture_cm, fov_deg и другие поля
func (h *Handler) PublishHandler(c *gin.Context) {
	idStr := c.PostForm("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}
	id := uint(id64)

	// Собираем обновления
	updates := map[string]interface{}{
		"description":     c.PostForm("description"),
		"aperture_cm":     atoiSafe(c.PostForm("aperture")),
		"fov_deg":         c.PostForm("fov"),
		"filters":         c.PostForm("filters"),
		"depth_mag":       c.PostForm("depth"),
		"time_resolution": c.PostForm("time_res"),
		"observatory":     c.PostForm("observatory"),
	}

	// Обновляем поля через ORM (метод репозитория)
	if err := h.Repo.UpdateFields(id, updates); err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка обновления")
		return
	}

	// Меняем статус на published
	if err := h.Repo.Publish(id); err != nil {
		logrus.Error(err)
		c.String(http.StatusInternalServerError, "Ошибка публикации")
		return
	}

	c.Redirect(http.StatusFound, "/grid")
}

// DeleteHandler – логическое удаление (POST /delete) через SQL UPDATE
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

// -------------------------------------------------------
// Вспомогательные функции
// -------------------------------------------------------

func atoiSafe(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
