package repository

import (
	"time"
	"user-service/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *UserRepository) Search(query string) ([]models.User, error) {
	var users []models.User
	q := "%" + query + "%"
	err := r.db.Where("firstname ILIKE ? OR lastname ILIKE ?", q, q).Find(&users).Error
	return users, err
}

func (r *UserRepository) UpdateUserRole(id uint, role models.Role) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}

func (r *UserRepository) UpdateUserBlock(id uint, until *time.Time) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("blocked_until", until).Error
}


