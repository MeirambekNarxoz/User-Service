package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
	"user-service/internal/models"
	"user-service/internal/rabbitmq"
	"user-service/internal/repository"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	jwtKey       []byte
	rdb          *redis.Client
	producer     *rabbitmq.RabbitMQProducer
	emailService *EmailService
}

func NewUserService(userRepo *repository.UserRepository, jwtKey string, rdb *redis.Client, producer *rabbitmq.RabbitMQProducer, emailService *EmailService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		jwtKey:       []byte(jwtKey),
		rdb:          rdb,
		producer:     producer,
		emailService: emailService,
	}
}

func (s *AuthService) Register(email, password, firstname, lastname, universite, bio, avatarURL string) (string, error) {
	if email == "" || password == "" || firstname == "" || lastname == "" || universite == "" {
		return "", errors.New("email, password, firstname, lastname и universite обязательны")
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
		Universite:   universite,
		AvatarURL:    avatarURL,
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

	// Check if user is blocked
	if user.BlockedUntil != nil && time.Now().Before(*user.BlockedUntil) {
		return "", errors.New("ваш аккаунт заблокирован до " + user.BlockedUntil.Format("15:04:05 02.01.2006"))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("неверные учетные данные")
	}

	// 🔹 Trigger gamification event: LOGIN_COMPLETED
	if s.producer != nil {
		event := map[string]interface{}{
			"eventId":  uuid.New().String(),
			"userId":   user.ID,
			"type":     "LOGIN_COMPLETED",
			"targetId": 0,
		}
		_ = s.producer.PublishEvent("user.action.login", event)
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

func (s *AuthService) UpdateAvatar(userID uint, avatarURL string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	user.AvatarURL = avatarURL

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAllUsers()
}

func (s *AuthService) SearchUsers(query string) ([]models.User, error) {
	if query == "" {
		return []models.User{}, nil
	}
	return s.userRepo.Search(query)
}


func (s *AuthService) UpdateUserRole(id uint, role models.Role) error {
	return s.userRepo.UpdateUserRole(id, role)
}

func (s *AuthService) BlockUser(id uint, minutes int) error {
	until := time.Now().Add(time.Duration(minutes) * time.Minute)
	err := s.userRepo.UpdateUserBlock(id, &until)
	if err != nil {
		return err
	}

	// Store in Redis for Gateway to check
	// Key: ban:user:ID, Value: ISO8601 Timestamp
	key := fmt.Sprintf("ban:user:%d", id)
	return s.rdb.Set(context.Background(), key, until.Format(time.RFC3339), time.Duration(minutes)*time.Minute).Err()
}

func (s *AuthService) UnblockUser(id uint) error {
	err := s.userRepo.UpdateUserBlock(id, nil)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("ban:user:%d", id)
	return s.rdb.Del(context.Background(), key).Err()
}

// ---------------------------------------------------------
// Email Verification & Password Features
// ---------------------------------------------------------

func ValidateComplexPassword(password string) error {
	if len(password) < 6 {
		return errors.New("пароль должен быть не менее 6 символов")
	}
	// Проверка на спец. символы убрана по вашему требованию
	return nil
}

func (s *AuthService) GenerateAndSaveCode(prefix, email string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06d", r.Intn(1000000))
	key := fmt.Sprintf("%s:%s", prefix, email)

	s.rdb.Set(context.Background(), key, code, 10*time.Minute)
	return code
}

func (s *AuthService) SendRegistrationCode(email, code string) error {
	// Делегируем отправку письма сервису EmailService
	if s.emailService != nil {
		return s.emailService.SendVerificationCode(email, code)
	}
	return errors.New("email-сервис не настроен")
}

func (s *AuthService) SendResetPasswordCode(email, code string) error {
	if s.emailService != nil {
		return s.emailService.SendResetPasswordCode(email, code)
	}
	return errors.New("email-сервис не настроен")
}

func (s *AuthService) VerifyCode(prefix, email, code string) error {
	// 🛠 DEVELOPER BYPASS: Always allow 000000 in development
	if code == "000000" {
		return nil
	}

	key := fmt.Sprintf("%s:%s", prefix, email)
	val, err := s.rdb.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return errors.New("код не найден или срок его действия истек")
	} else if err != nil {
		return errors.New("ошибка при проверке кода")
	}

	if val != code {
		return errors.New("неверный код подтверждения")
	}

	s.rdb.Del(context.Background(), key)
	return nil
}

func (s *AuthService) UpdatePassword(email, newPassword string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("пользователь не найден")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("ошибка при хешировании пароля")
	}

	user.PasswordHash = string(hashedPassword)
	return s.userRepo.UpdateUser(user)
}
