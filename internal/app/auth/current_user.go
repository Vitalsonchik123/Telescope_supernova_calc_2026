package auth

import (
	"errors"
	"supernova-calc/internal/app/models"
	"supernova-calc/internal/app/repository"
)

// currentUserID — ID пользователя, зафиксированный на всю ЛР3.
// В ЛР4 заменим на реальную авторизацию через сессии/JWT.
const currentUserID uint = 1

var currentRepo *repository.Repository

// InitCurrentUser — вызывается один раз при старте приложения,
// чтобы установить «пользователя по умолчанию» (singleton).
func InitCurrentUser(repo *repository.Repository) {
	currentRepo = repo
}

// GetCurrentUserID — возвращает ID текущего пользователя.
// Используется во всех API-методах, где нужно знать создателя.
func GetCurrentUserID() uint {
	return currentUserID
}

// GetCurrentUser — возвращает модель текущего пользователя из БД.
func GetCurrentUser() (*models.User, error) {
	if currentRepo == nil {
		return nil, errors.New("current user not initialized")
	}
	return currentRepo.GetUserByID(currentUserID)
}
