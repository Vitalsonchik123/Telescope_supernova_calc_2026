package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"supernova-calc/internal/app/auth"
	"supernova-calc/internal/app/models"
)

func (h *Handler) apiError(c *gin.Context, code int, msg string) {
	c.JSON(code, models.APIError{Status: "fail", Message: msg})
}

// ================================================================
// GET /api/telescopes
// ================================================================

// APIGetTelescopes godoc
// @Summary      Список опубликованных телескопов
// @Description  Возвращает список опубликованных услуг с фильтром по диаметру и автору
// @Tags         telescopes
// @Produce      json
// @Param        min_aperture  query     int   false  "Минимальный диаметр (см)"
// @Param        is_mine       query     bool  false  "Только мои услуги"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  models.APIError
// @Router       /telescopes [get]
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

	// Гость может смотреть список. Если авторизован — берём user_id из JWT.
	currentUserID, _ := auth.GetUserIDFromContext(c) // 0, если гость

	var result []models.TelescopeListSerializer
	for _, t := range telescopes {
		likes, _ := h.Repo.GetLikesCount(t.ID)
		isMine := currentUserID != 0 && t.UserID == currentUserID
		isLiked := false
		if currentUserID != 0 {
			isLiked, _ = h.Repo.HasUserLiked(currentUserID, t.ID)
		}

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

// ================================================================
// GET /api/feed
// ================================================================

// APIFeed godoc
// @Summary      Лента опубликованных услуг
// @Description  Возвращает первый опубликованный телескоп или следующий по ID
// @Tags         telescopes
// @Produce      json
// @Param        id    query     int   false  "ID текущего телескопа"
// @Param        next  query     bool  false  "Перейти к следующему"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  models.APIError
// @Failure      404  {object}  models.APIError
// @Router       /feed [get]
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

	currentUserID, _ := auth.GetUserIDFromContext(c) // 0, если гость
	isLiked := false
	if currentUserID != 0 {
		isLiked, _ = h.Repo.HasUserLiked(currentUserID, t.ID)
	}

	full := models.TelescopeFullSerializer{
		Telescope: t,
		Creator: models.UserSerializer{
			ID:   creator.ID,
			Name: creator.Name,
			Role: creator.Role,
		},
		ImageURL:   h.Repo.GetMinioURL(t.ImageKey),
		VideoURL:   h.Repo.GetMinioURL(t.VideoKey),
		LikesCount: int(likes),
		IsMine:     currentUserID != 0 && t.UserID == currentUserID,
		IsLiked:    isLiked,
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": full})
}

// ================================================================
// GET /api/draft — защищённый
// ================================================================

// APIGetDraft godoc
// @Summary      Получить черновик текущего пользователя
// @Description  Возвращает единственный черновик пользователя (не более 1)
// @Tags         telescopes
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /draft [get]
func (h *Handler) APIGetDraft(c *gin.Context) {
	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		h.apiError(c, http.StatusUnauthorized, "Не авторизован")
		return
	}

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

// ================================================================
// POST /api/telescopes — защищённый
// ================================================================

// APICreateTelescope godoc
// @Summary      Создать черновик услуги
// @Description  Создаёт черновик и загружает картинку + видео в MinIO
// @Tags         telescopes
// @Accept       multipart/form-data
// @Produce      json
// @Param        name   formData  string  true  "Название услуги"
// @Param        image  formData  file    false "Изображение"
// @Param        video  formData  file    false "Короткое видео"
// @Success      201  {object}  map[string]interface{}
// @Failure      401  {object}  models.APIError
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /telescopes [post]
func (h *Handler) APICreateTelescope(c *gin.Context) {
	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		h.apiError(c, http.StatusUnauthorized, "Не авторизован")
		return
	}

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

// ================================================================
// PUT /api/telescopes/:id/publish — защищённый
// ================================================================

// APIPublishTelescope godoc
// @Summary      Опубликовать услугу
// @Description  Меняет статус черновика на published. Только для автора.
// @Tags         telescopes
// @Accept       json
// @Produce      json
// @Param        id    path      int                           true  "ID услуги"
// @Param        body  body      models.PublishTelescopeRequest true  "Данные для публикации"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  models.APIError
// @Failure      401  {object}  models.APIError
// @Failure      403  {object}  models.APIError
// @Failure      404  {object}  models.APIError
// @Router       /telescopes/{id}/publish [put]
func (h *Handler) APIPublishTelescope(c *gin.Context) {
	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		h.apiError(c, http.StatusUnauthorized, "Не авторизован")
		return
	}

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
	if tel.UserID != userID {
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

// ================================================================
// DELETE /api/telescopes/:id — защищённый
// ================================================================

// APIDeleteTelescope godoc
// @Summary      Логическое удаление услуги
// @Description  Меняет статус на deleted. Только для автора.
// @Tags         telescopes
// @Produce      json
// @Param        id  path      int  true  "ID услуги"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  models.APIError
// @Failure      401  {object}  models.APIError
// @Failure      403  {object}  models.APIError
// @Failure      404  {object}  models.APIError
// @Router       /telescopes/{id} [delete]
func (h *Handler) APIDeleteTelescope(c *gin.Context) {
	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		h.apiError(c, http.StatusUnauthorized, "Не авторизован")
		return
	}

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
	if tel.UserID != userID {
		h.apiError(c, http.StatusForbidden, "Можно удалять только свои услуги")
		return
	}

	if err := h.Repo.SoftDeleteTelescope(id); err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка удаления")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Услуга удалена"})
}

// ================================================================
// POST /api/telescopes/:id/like — защищённый
// ================================================================

// APILikeTelescope godoc
// @Summary      Поставить или снять лайк
// @Description  Принимает like=1 (поставить) или like=0 (снять)
// @Tags         telescopes
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "ID услуги"
// @Param        body  body      models.LikeRequest   true  "Данные лайка"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  models.APIError
// @Failure      401  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /telescopes/{id}/like [post]
func (h *Handler) APILikeTelescope(c *gin.Context) {
	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		h.apiError(c, http.StatusUnauthorized, "Не авторизован")
		return
	}

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

	if err := h.Repo.SetLike(userID, uint(id64), req.Like); err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка лайка")
		return
	}

	count, _ := h.Repo.GetLikesCount(uint(id64))
	c.JSON(http.StatusOK, gin.H{"status": "success", "likes_count": count})
}

// ================================================================
// POST /api/users/register
// ================================================================

// APIRegister godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя с хешированным паролем
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      models.RegisterRequest  true  "Данные регистрации"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /users/register [post]
func (h *Handler) APIRegister(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный формат: "+err.Error())
		return
	}

	existing, _ := h.Repo.GetUserByName(req.Name)
	if existing != nil {
		h.apiError(c, http.StatusBadRequest, "Пользователь с таким именем уже существует")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка хеширования пароля")
		return
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user, err := h.Repo.CreateUser(req.Name, string(hash), role)
	if err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка регистрации")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": models.UserSerializer{
			ID:   user.ID,
			Name: user.Name,
			Role: user.Role,
		},
	})
}

// ================================================================
// POST /api/users/login
// ================================================================

// APILogin godoc
// @Summary      Аутентификация
// @Description  Возвращает JWT-токен при успешном входе
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      models.LoginRequest  true  "Данные для входа"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  models.APIError
// @Failure      401  {object}  models.APIError
// @Router       /users/login [post]
func (h *Handler) APILogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.apiError(c, http.StatusBadRequest, "Неверный формат")
		return
	}

	user, err := h.Repo.GetUserByName(req.Name)
	if err != nil {
		h.apiError(c, http.StatusUnauthorized, "Неверное имя или пароль")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.apiError(c, http.StatusUnauthorized, "Неверное имя или пароль")
		return
	}

	token, err := auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": models.LoginResponse{
			Token: token,
			User: models.UserSerializer{
				ID:   user.ID,
				Name: user.Name,
				Role: user.Role,
			},
		},
	})
}

// ================================================================
// POST /api/users/logout
// ================================================================

// APILogout godoc
// @Summary      Деавторизация
// @Description  Добавляет JWT в blacklist Redis. Токен больше не действителен.
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /users/logout [post]
func (h *Handler) APILogout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		h.apiError(c, http.StatusUnauthorized, "Отсутствует заголовок Authorization")
		return
	}

	tokenString := strings.TrimPrefix(header, "Bearer ")

	claims, err := auth.ParseJWT(tokenString)
	if err != nil {
		h.apiError(c, http.StatusUnauthorized, "Неверный токен: "+err.Error())
		return
	}

	var ttl time.Duration
	if claims.ExpiresAt != nil {
		ttl = time.Until(claims.ExpiresAt.Time)
		if ttl < 0 {
			ttl = 0
		}
	} else {
		ttl = 24 * time.Hour
	}

	if err := h.Redis.WriteJWTToBlacklist(c.Request.Context(), tokenString, ttl); err != nil {
		h.apiError(c, http.StatusInternalServerError, "Ошибка logout: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Вы вышли из системы",
	})
}
