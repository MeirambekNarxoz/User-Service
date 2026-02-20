package routes

import (
	"github.com/gin-gonic/gin"
	"user-service/internal/delivery/http" // Импорт хандлера
)

func MapUserRoutes(r *gin.Engine, h *http.AuthHandler) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}
