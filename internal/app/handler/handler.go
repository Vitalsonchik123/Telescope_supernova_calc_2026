package handler

import (
    "net/http"
    "strconv"
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

func getMinioURL(key string) string {
    return "http://localhost:9000/supernova-data/" + key
}

// FeedHandler – лента по ID (GET /feed?id=xxx&next=true)
func (h *Handler) FeedHandler(c *gin.Context) {
    id := c.Query("id")
    next := c.Query("next") == "true"

    if id == "" {
        c.String(http.StatusBadRequest, "Не указан ID телескопа")
        return
    }

    tel, err := h.Repo.GetByID(id)
    if err != nil {
        logrus.Error(err)
        c.String(http.StatusNotFound, "Телескоп не найден")
        return
    }

    if next {
        all := h.Repo.GetAll()
        for i, t := range all {
            if t.ID == id && i+1 < len(all) {
                c.Redirect(http.StatusFound, "/feed?id="+all[i+1].ID)
                return
            }
        }
        // если следующего нет, зацикливаем на первый
        if len(all) > 0 {
            c.Redirect(http.StatusFound, "/feed?id="+all[0].ID)
        } else {
            c.String(http.StatusNotFound, "Нет телескопов")
        }
        return
    }

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
    draft, err := h.Repo.GetDraft()
    if err != nil {
        logrus.Error(err)
        c.String(http.StatusNotFound, "Черновик не найден")
        return
    }

    view := models.TelescopeView{
        Telescope: draft,
        ImageURL:  getMinioURL(draft.ImageKey),
        VideoURL:  getMinioURL(draft.VideoKey),
        Likes:     len(draft.LikedBy),
    }

    c.HTML(http.StatusOK, "add.html", gin.H{
        "telescope": view,
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
