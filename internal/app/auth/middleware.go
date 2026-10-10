package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	redisclient "supernova-calc/internal/app/redis"
)

const jwtPrefix = "Bearer "

// ContextUserID — ключ для user_id в контексте Gin
const ContextUserID = "user_id"

// ContextUserRole — ключ для роли в контексте Gin
const ContextUserRole = "user_role"

// RequireAuth — middleware, которая требует валидный JWT
// Если токена нет / он просрочен / в blacklist — возвращает 401
func RequireAuth(redis *redisclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Получаем заголовок
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, jwtPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "fail",
				"message": "Отсутствует или неверный заголовок Authorization",
			})
			return
		}

		tokenString := strings.TrimPrefix(header, jwtPrefix)

		// 2. Проверяем blacklist
		if redis != nil && redis.CheckJWTInBlacklist(context.Background(), tokenString) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "fail",
				"message": "Токен отозван",
			})
			return
		}

		// 3. Парсим и валидируем JWT
		claims, err := ParseJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "fail",
				"message": "Неверный токен: " + err.Error(),
			})
			return
		}

		// 4. Кладём данные пользователя в контекст
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserRole, claims.Role)

		c.Next()
	}
}

// GetUserIDFromContext — читает user_id из контекста (если был установлен middleware)
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	val, exists := c.Get(ContextUserID)
	if !exists {
		return 0, false
	}
	id, ok := val.(uint)
	return id, ok
}

// GetUserRoleFromContext — читает роль из контекста
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextUserRole)
	if !exists {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}
