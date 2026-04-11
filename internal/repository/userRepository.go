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

// Friendship Methods

func (r *UserRepository) CreateFriendRequest(friendship *models.Friendship) error {
	return r.db.Create(friendship).Error
}

func (r *UserRepository) UpdateFriendStatus(userID, friendID uint, status models.FriendshipStatus) error {
	return r.db.Model(&models.Friendship{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userID, friendID, friendID, userID).
		Update("status", status).Error
}

func (r *UserRepository) GetFriendships(userID uint, status models.FriendshipStatus) ([]models.Friendship, error) {
	var friendships []models.Friendship
	err := r.db.Preload("Friend").Preload("User").
		Where("(user_id = ? OR friend_id = ?) AND status = ?", userID, userID, status).
		Find(&friendships).Error
	return friendships, err
}

func (r *UserRepository) GetFriendshipBetween(user1ID, user2ID uint) (*models.Friendship, error) {
	var f models.Friendship
	err := r.db.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", user1ID, user2ID, user2ID, user1ID).First(&f).Error
	return &f, err
}
