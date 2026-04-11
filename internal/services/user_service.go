package services

import (
	"context"
	"errors"
	"fmt"
	"time"
	"user-service/internal/models"
	"user-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtKey   []byte
	rdb      *redis.Client
}

func NewUserService(userRepo *repository.UserRepository, jwtKey string, rdb *redis.Client) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtKey:   []byte(jwtKey),
		rdb:      rdb,
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

func (s *AuthService) SendFriendRequest(userID, friendID uint) error {
	if userID == friendID {
		return errors.New("нельзя добавить самого себя в друзья")
	}

	// Check if already exists
	f, err := s.userRepo.GetFriendshipBetween(userID, friendID)
	if err == nil {
		if f.Status == models.FriendshipAccepted {
			return errors.New("вы уже друзья")
		}
		return errors.New("запрос уже отправлен или ожидает подтверждения")
	}

	friendship := &models.Friendship{
		UserID:   userID,
		FriendID: friendID,
		Status:   models.FriendshipPending,
	}

	return s.userRepo.CreateFriendRequest(friendship)
}

func (s *AuthService) AcceptFriendRequest(userID, friendID uint) error {
	return s.userRepo.UpdateFriendStatus(userID, friendID, models.FriendshipAccepted)
}

func (s *AuthService) GetFriends(userID uint) ([]models.User, error) {
	friendships, err := s.userRepo.GetFriendships(userID, models.FriendshipAccepted)
	if err != nil {
		return nil, err
	}

	var friends []models.User
	for _, f := range friendships {
		if f.UserID == userID {
			friends = append(friends, f.Friend)
		} else {
			friends = append(friends, f.User)
		}
	}
	return friends, nil
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
