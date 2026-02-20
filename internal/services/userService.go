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

// Register — регистрация нового пользователя со всеми твоими полями
func (s *AuthService) Register(user *models.User) (string, error) {
	// 1. Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("ошибка при хешировании пароля")
	}
	user.Password = string(hashedPassword)

	// Проверка обязательных полей (можно расширить)
	if user.Email == "" {
		return "", errors.New("email не может быть пустым")
	}

	// 2. Сохранение в базу данных
	err = s.userRepo.CreateUser(user)
	if err != nil {
		return "", errors.New("не удалось создать пользователя (возможно, email уже занят)")
	}

	// 3. Генерация токена сразу после регистрации
	token, err := s.GenerateJwtToken(user)
	if err != nil {
		return "", errors.New("ошибка генерации токена")
	}

	return token, nil
}

// Login — вход по Email и паролю
func (s *AuthService) Login(email, password string) (string, error) {
	// 1. Поиск пользователя по Email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", errors.New("неверные учетные данные")
	}

	// 2. Проверка пароля через Bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("неверные учетные данные")
	}

	// 3. Генерация JWT-токена
	token, err := s.GenerateJwtToken(user)
	if err != nil {
		return "", errors.New("ошибка при создании токена")
	}

	return token, nil
}

// GenerateJwtToken — создание токена с данными твоей модели
func (s *AuthService) GenerateJwtToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":      user.Email,
		"firstname":  user.Firstname,
		"role":       user.Role,
		"status":     user.Status, // Студент/Школьник и т.д.
		"universite": user.Universite,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // 24 часа
	})

	signedToken, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
