package services

import (
	"errors"
	"time"
	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtKey   []byte
}

func NewUserService(userRepo *repository.UserRepository, jwtKey string) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtKey:   []byte(jwtKey),
	}
}

func (s *AuthService) Register(email, password, firstname, lastname, bio string) (string, error) {
	if email == "" || password == "" || firstname == "" || lastname == "" {
		return "", errors.New("email, password, firstname и lastname обязательны")
	}

	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		return "", errors.New("email уже занят")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("ошибка при хешировании пароля")
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Firstname:    firstname,
		Lastname:     lastname,
		Bio:          bio,
		Role:         models.RoleUser,
		Status:       models.StatusActive,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return "", errors.New("не удалось создать пользователя")
	}

	return s.GenerateJwtToken(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", errors.New("неверные учетные данные")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("неверные учетные данные")
	}

	return s.GenerateJwtToken(user)
}

func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *AuthService) UpdateProfile(userID uint, firstname, lastname, bio string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if firstname != "" {
		user.Firstname = firstname
	}
	if lastname != "" {
		user.Lastname = lastname
	}
	user.Bio = bio

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) GenerateJwtToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"firstname": user.Firstname,
		"lastname":  user.Lastname,
		"role":      user.Role,
		"status":    user.Status,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(s.jwtKey)
}

func (s *AuthService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAllUsers()
}
