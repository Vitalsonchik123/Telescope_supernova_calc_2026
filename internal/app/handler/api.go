package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"supernova-calc/internal/app/auth"
	"supernova-calc/internal/app/models"
)

func (h *Handler) apiError(c *gin.Context, code int, msg string) {
	c.JSON(code, models.APIError{Status: "fail", Message: msg})
}

// GET /api/telescopes?min_aperture=100&is_mine=true
func (h *Handler) APIGetTelescopes(c *gin.Context) {
	minStr := c.Query("min_aperture")
	minAperture := 0
	if minStr != "" {
		if v, err := strconv.Atoi(minStr); err == nil {
			minAperture = v
		}
	}
	isMineFilter := c.Query("is_mine") == "true"

	telescopes, err := h.Repo.GetPublishedTelescopes(minAperture)
	if err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка БД")
		return
	}

	currentUserID := auth.GetCurrentUserID()

	var result []models.TelescopeListSerializer
	for _, t := range telescopes {
		likes, _ := h.Repo.GetLikesCount(t.ID)
		isMine := t.UserID == currentUserID
		isLiked, _ := h.Repo.HasUserLiked(currentUserID, t.ID)

		if isMineFilter && !isMine {
			continue
		}
		result = append(result, models.TelescopeListSerializer{
			ID:         t.ID,
			Name:       t.Name,
			ApertureCm: t.ApertureCm,
			FovDeg:     t.FovDeg,
			ImageURL:   h.Repo.GetMinioURL(t.ImageKey),
			LikesCount: int(likes),
			IsMine:     isMine,
			IsLiked:    isLiked,
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": result})
}

// GET /api/feed?id=1&next=true
func (h *Handler) APIFeed(c *gin.Context) {
	idStr := c.Query("id")
	next := c.Query("next") == "true"

	if idStr == "" {
		all, err := h.Repo.GetPublishedTelescopes(0)
		if err != nil || len(all) == 0 {
			h.apiError(c, http.StatusNotFound, "Нет услуг")
			return
		}
		idx := 0
		if next {
			idx = 1 % len(all)
		}
		h.renderFeed(c, all[idx])
		return
	}

	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный ID")
		return
	}
	id := uint(id64)

	tel, err := h.Repo.GetByID(id)
	if err != nil {
		h.apiError(c, http.StatusNotFound, "Услуга не найдена")
		return
	}

	if next {
		all, _ := h.Repo.GetPublishedTelescopes(0)
		currentIndex := -1
		for i, t := range all {
			if t.ID == id {
				currentIndex = i
				break
			}
		}
		if currentIndex == -1 {
			h.apiError(c, http.StatusNotFound, "Услуга не найдена")
			return
		}
		ni := (currentIndex + 1) % len(all)
		h.renderFeed(c, all[ni])
		return
	}

	h.renderFeed(c, *tel)
}

func (h *Handler) renderFeed(c *gin.Context, t models.Telescope) {
	likes, _ := h.Repo.GetLikesCount(t.ID)
	creator, _ := h.Repo.GetUserByID(t.UserID)
	currentUserID := auth.GetCurrentUserID()
	isLiked, _ := h.Repo.HasUserLiked(currentUserID, t.ID)

	full := models.TelescopeFullSerializer{
		Telescope:  t,
		Creator:    models.UserSerializer{ID: creator.ID, Name: creator.Name},
		ImageURL:   h.Repo.GetMinioURL(t.ImageKey),
		VideoURL:   h.Repo.GetMinioURL(t.VideoKey),
		LikesCount: int(likes),
		IsMine:     t.UserID == currentUserID,
		IsLiked:    isLiked,
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": full})
}

// GET /api/draft
func (h *Handler) APIGetDraft(c *gin.Context) {
	userID := auth.GetCurrentUserID()
	draft, err := h.Repo.GetDraft(userID)
	if err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка БД")
		return
	}
	if draft == nil {
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": draft})
}

// POST /api/telescopes
func (h *Handler) APICreateTelescope(c *gin.Context) {
	userID := auth.GetCurrentUserID()

	existing, _ := h.Repo.GetDraft(userID)
	if existing != nil {
		h.apiError(c, http.StatusBadRequest, "У пользователя уже есть черновик")
		return
	}

	if err := c.Request.ParseMultipartForm(20 << 20); err != nil {
		h.apiError(c, http.StatusBadRequest, "Ошибка чтения формы")
		return
	}

	name := c.Request.FormValue("name")
	if name == "" {
		h.apiError(c, http.StatusBadRequest, "Не указано название")
		return
	}

	t := &models.Telescope{
		Name:   name,
		UserID: userID,
		Status: "draft",
	}
	if err := h.Repo.CreateDraft(t); err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка создания")
		return
	}

	if header, err := c.FormFile("image"); err == nil {
		filename, err := h.Repo.UploadFile("img", t.ID, header)
		if err == nil {
			_ = h.Repo.AddTelescopeImage(t.ID, "image_key", filename)
		}
	}
	if header, err := c.FormFile("video"); err == nil {
		filename, err := h.Repo.UploadFile("vid", t.ID, header)
		if err == nil {
			_ = h.Repo.AddTelescopeImage(t.ID, "video_key", filename)
		}
	}

	updated, _ := h.Repo.GetByIDAny(t.ID)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": updated})
}

// PUT /api/telescopes/:id/publish
func (h *Handler) APIPublishTelescope(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный ID")
		return
	}
	id := uint(id64)

	tel, err := h.Repo.GetByIDAny(id)
	if err != nil {
		h.apiError(c, http.StatusNotFound, "Услуга не найдена")
		return
	}
	if tel.UserID != auth.GetCurrentUserID() {
		h.apiError(c, http.StatusForbidden, "Можно публиковать только свои услуги")
		return
	}

	var req models.PublishTelescopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный формат: "+err.Error())
		return
	}

	updates := map[string]interface{}{
		"description": req.Description,
		"aperture_cm": req.ApertureCm,
		"fov_deg":     req.FovDeg,
	}
	if err := h.Repo.PublishTelescope(id, updates); err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка публикации")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Опубликовано"})
}

// DELETE /api/telescopes/:id
func (h *Handler) APIDeleteTelescope(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный ID")
		return
	}
	id := uint(id64)

	tel, err := h.Repo.GetByIDAny(id)
	if err != nil {
		h.apiError(c, http.StatusNotFound, "Услуга не найдена")
		return
	}
	if tel.UserID != auth.GetCurrentUserID() {
		h.apiError(c, http.StatusForbidden, "Можно удалять только свои услуги")
		return
	}

	if err := h.Repo.SoftDeleteTelescope(id); err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка удаления")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Услуга удалена"})
}

// POST /api/telescopes/:id/like
func (h *Handler) APILikeTelescope(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный ID")
		return
	}

	var req models.LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный формат")
		return
	}
	if req.Like != 0 && req.Like != 1 {
		h.apiError(c, http.StatusBadRequest, "Поле like должно быть 0 или 1")
		return
	}

	userID := auth.GetCurrentUserID()
	if err := h.Repo.SetLike(userID, uint(id64), req.Like); err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка лайка")
		return
	}

	count, _ := h.Repo.GetLikesCount(uint(id64))
	c.JSON(http.StatusOK, gin.H{"status": "success", "likes_count": count})
}

// POST /api/users/register
func (h *Handler) APIRegister(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный формат: "+err.Error())
		return
	}

	user, err := h.Repo.RegisterUser(req.Name)
	if err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка регистрации")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   models.UserSerializer{ID: user.ID, Name: user.Name},
	})
}

func (h *Handler) APILogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный формат")
		return
	}

	user, err := h.Repo.GetUserByName(req.Name)
	if err != nil {
		h.apiError(c, http.StatusUnauthorized, "Пользователь не найден")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Заглушка авторизации. В ЛР4 здесь будет JWT",
		"data":    models.UserSerializer{ID: user.ID, Name: user.Name},
	})
}

func (h *Handler) APILogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Заглушка деавторизации. В ЛР4 здесь будет JWT",
	})
}
