package repository

import (
	"errors"

	"gorm.io/gorm"

	"supernova-calc/internal/app/models"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetAll – возвращает только опубликованные услуги
func (r *Repository) GetAll() ([]models.Telescope, error) {
	var telescopes []models.Telescope
	err := r.db.
		Where("status = ?", "published").
		Preload("User").
		Find(&telescopes).Error
	return telescopes, err
}

// GetByID – возвращает услугу по ID, если она опубликована
func (r *Repository) GetByID(id uint) (*models.Telescope, error) {
	var telescope models.Telescope
	err := r.db.
		Where("id = ? AND status = ?", id, "published").
		Preload("User").
		First(&telescope).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("телескоп не найден")
	}
	return &telescope, err
}

// GetDraft – черновик пользователя
func (r *Repository) GetDraft(userID uint) (*models.Telescope, error) {
	var telescope models.Telescope
	err := r.db.
		Where("user_id = ? AND status = ?", userID, "draft").
		First(&telescope).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &telescope, err
}

// CreateDraft – создаёт новую услугу со статусом draft
func (r *Repository) CreateDraft(t *models.Telescope) error {
	t.Status = "draft"
	return r.db.Create(t).Error
}

// Publish – меняет статус на published
func (r *Repository) Publish(id uint) error {
	return r.db.
		Model(&models.Telescope{}).
		Where("id = ?", id).
		Update("status", "published").Error
}

// Delete – логическое удаление через SQL UPDATE (без ORM)
func (r *Repository) Delete(id uint) error {
	sql := "UPDATE telescopes SET status = 'deleted' WHERE id = $1"
	return r.db.Exec(sql, id).Error
}

// FilterByApertureMin – фильтрация только по опубликованным
func (r *Repository) FilterByApertureMin(minAperture int) ([]models.Telescope, error) {
	var telescopes []models.Telescope
	err := r.db.
		Where("status = ? AND aperture_cm >= ?", "published", minAperture).
		Preload("User").
		Find(&telescopes).Error
	return telescopes, err
}

// GetLikesCount – количество лайков
func (r *Repository) GetLikesCount(telescopeID uint) (int64, error) {
	var count int64
	err := r.db.
		Model(&models.Like{}).
		Where("telescope_id = ?", telescopeID).
		Count(&count).Error
	return count, err
}

// UpdateFields – обновляет поля через ORM
func (r *Repository) UpdateFields(id uint, updates map[string]interface{}) error {
	return r.db.
		Model(&models.Telescope{}).
		Where("id = ?", id).
		Updates(updates).Error
}
