package http

import (
	"net/http"
	"strconv"
	"user-service/internal/models"
	"user-service/internal/services"
	"user-service/internal/storage"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
	minioClient *storage.MinioClient
}

type RegisterRequest struct {
	Email      string `json:"email" form:"email" binding:"required"`
	Password   string `json:"password" form:"password" binding:"required"`
	Firstname  string `json:"firstname" form:"firstname" binding:"required"`
	Lastname   string `json:"lastname" form:"lastname" binding:"required"`
	Universite string `json:"universite" form:"universite" binding:"required"`
	Bio        string `json:"bio" form:"bio"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Bio       string `json:"bio"`
}

func NewAuthHandler(s *services.AuthService, minio *storage.MinioClient) *AuthHandler {
	return &AuthHandler{
		authService: s,
		minioClient: minio,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	// 🔹 Биндим form-data (важно: не JSON)
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "неверные данные: проверь обязательные поля",
		})
		return
	}

	// 🔹 Валидация (дополнительно, чтобы не надеяться только на binding)
	if req.Email == "" || req.Password == "" || req.Firstname == "" || req.Lastname == "" || req.Universite == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "все обязательные поля (email, пароль, имя, фамилия, университет) должны быть заполнены",
		})
		return
	}

	// 🔹 Работа с файлом
	var avatarURL string
	file, err := c.FormFile("avatar")
	if err == nil {
		avatarURL, err = h.minioClient.UploadAvatar(c.Request.Context(), file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "ошибка при загрузке аватара: " + err.Error(),
			})
			return
		}
	}

	// 🔹 Регистрация
	token, err := h.authService.Register(
		req.Email,
		req.Password,
		req.Firstname,
		req.Lastname,
		req.Universite,
		req.Bio,
		avatarURL,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 🔹 Ответ
	c.JSON(http.StatusCreated, gin.H{
		"token":      token,
		"avatar_url": avatarURL,
	})
}
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email и пароль обязательны"})
		return
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	user, err := h.authService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные данные"})
		return
	}

	user, err := h.authService.UpdateProfile(
		uint(id),
		req.Firstname,
		req.Lastname,
		req.Bio,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "профиль обновлен",
		"user":    user,
	})
}

func (h *AuthHandler) GetAllUsers(c *gin.Context) {
	users, err := h.authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при получении пользователей"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл аватара не найден"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	// Upload to MinIO
	avatarURL, err := h.minioClient.UploadAvatar(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при загрузке: " + err.Error()})
		return
	}

	// Update user profile
	user, err := h.authService.UpdateAvatar(userID, avatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при обновлении профиля"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "аватар обновлен",
		"avatar_url": avatarURL,
		"user":       user,
	})
}

func (h *AuthHandler) SearchUsers(c *gin.Context) {
	query := c.Query("query")
	users, err := h.authService.SearchUsers(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при поиске"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Friendship Handlers

func (h *AuthHandler) SendFriendRequest(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var req struct {
		FriendID uint `json:"friend_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}

	if err := h.authService.SendFriendRequest(userID, req.FriendID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "заявка отправлена"})
}

func (h *AuthHandler) AcceptFriendRequest(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var req struct {
		FriendID uint `json:"friend_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
		return
	}

	if err := h.authService.AcceptFriendRequest(userID, req.FriendID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при подтверждении"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "заявка принята"})
}

func (h *AuthHandler) GetFriends(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	friends, err := h.authService.GetFriends(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при получении списка"})
		return
	}
	c.JSON(http.StatusOK, friends)
}

func (h *AuthHandler) UpdateRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректная роль"})
		return
	}

	if err := h.authService.UpdateUserRole(uint(id), models.Role(req.Role)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при обновлении роли"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "роль обновлена"})
}

func (h *AuthHandler) BlockUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		Minutes int `json:"minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректное время"})
		return
	}

	if err := h.authService.BlockUser(uint(id), req.Minutes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при блокировке"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "пользователь заблокирован"})
}

func (h *AuthHandler) UnblockUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.authService.UnblockUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при разблокировке"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "пользователь разблокирован"})
}
