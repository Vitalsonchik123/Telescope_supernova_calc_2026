package repository

import (
	"errors"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"

	"supernova-calc/internal/app/models"
)

type Repository struct {
	db          *gorm.DB
	minio       *minio.Client
	minioBucket string
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

// ================================================================
// ЛР 3
// ================================================================

// GetUserByID — для singleton-пользователя
func (r *Repository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetPublishedTelescopes — список с фильтром по диаметру
func (r *Repository) GetPublishedTelescopes(minAperture int) ([]models.Telescope, error) {
	var telescopes []models.Telescope
	q := r.db.Where("status = ?", "published")
	if minAperture > 0 {
		q = q.Where("aperture_cm >= ?", minAperture)
	}
	err := q.Order("id").Find(&telescopes).Error
	return telescopes, err
}

// AddTelescopeImage — сохраняет путь к файлу в MinIO (в поле image_key или video_key)
func (r *Repository) AddTelescopeImage(id uint, field string, filename string) error {
	return r.db.Model(&models.Telescope{}).Where("id = ?", id).Update(field, filename).Error
}

// PublishTelescope — смена статуса на published + обновление полей
func (r *Repository) PublishTelescope(id uint, updates map[string]interface{}) error {
	updates["status"] = "published"
	return r.db.Model(&models.Telescope{}).Where("id = ?", id).Updates(updates).Error
}

// SoftDeleteTelescope — логическое удаление
func (r *Repository) SoftDeleteTelescope(id uint) error {
	return r.db.Model(&models.Telescope{}).Where("id = ?", id).Update("status", "deleted").Error
}

// SetLike — поставить/снять лайк
func (r *Repository) SetLike(userID, telescopeID uint, like int) error {
	if like == 1 {
		// Создаём, если не существует
		var existing models.Like
		err := r.db.Where("user_id = ? AND telescope_id = ?", userID, telescopeID).First(&existing).Error
		if err == nil {
			return nil // уже есть
		}
		newLike := models.Like{UserID: userID, TelescopeID: telescopeID}
		return r.db.Create(&newLike).Error
	}
	// like == 0 — удаляем
	return r.db.Where("user_id = ? AND telescope_id = ?", userID, telescopeID).Delete(&models.Like{}).Error
}

// RegisterUser — POST /api/users/register
func (r *Repository) RegisterUser(name string) (*models.User, error) {
	user := models.User{Name: name}
	if err := r.db.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByName — для логина (заглушка)
func (r *Repository) GetUserByName(name string) (*models.User, error) {
	var user models.User
	err := r.db.Where("name = ?", name).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
