package services

import (
	"errors"
	"time"
	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtKey   []byte
}

func NewUserService(userRepo *repository.UserRepository, jwtKey string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtKey: []byte(jwtKey)}
}

func (s *AuthService) Register(user *models.User) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("ошибка при хешировании пароля")
	}
	user.Password = string(hashedPassword)

	if user.Email == "" {
		return "", errors.New("email не может быть пустым")
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		return "", errors.New("не удалось создать пользователя (возможно, email уже занят)")
	}

	return s.GenerateJwtToken(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", errors.New("неверные учетные данные")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("неверные учетные данные")
	}

	return s.GenerateJwtToken(user)
}

// Получение пользователя по ID
func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// Обновление профиля
func (s *AuthService) UpdateProfile(userID uint, updatedData *models.User) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	// Обновляем поля
	user.Firstname = updatedData.Firstname
	user.Universite = updatedData.Universite
	user.Status = updatedData.Status

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) GenerateJwtToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID":         user.ID,
		"email":      user.Email,
		"firstname":  user.Firstname,
		"role":       user.Role,
		"status":     user.Status,
		"universite": user.Universite,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(s.jwtKey)
}
