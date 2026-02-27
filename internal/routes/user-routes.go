package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"user-service/internal/delivery/http"
	"user-service/internal/repository"
	"user-service/internal/services"
)

func SetupRouter(db *gorm.DB, router *gin.Engine) *gin.Engine {
	// 1. Инициализация слоев (Dependency Injection)
	// Замени "your_secret_key" на свой ключ из конфига/окружения
	userRepo := repository.NewUserRepository(db)
	authService := services.NewUserService(userRepo, "Aaa123")
	authHandler := http.NewAuthHandler(authService)

	// 2. Настройка групп и роутов
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)

		// Роуты для работы с профилем по ID
		auth.GET("/user/:id", authHandler.GetUser)
		auth.PUT("/update/:id", authHandler.UpdateProfile)
	}

	return router
}
