package main

import (
	"github.com/gin-gonic/gin"
	"user-service/internal/config"
	"user-service/internal/database"
	"user-service/internal/delivery/http"
	"user-service/internal/repository"
	"user-service/internal/services"
)

func main() {
	// 1. Загружаем конфиг (обращаемся через имя пакета config)
	cfg := config.LoadConfig()

	// 2. Инициализируем БД (передаем поле из структуры cfg)
	db := database.InitDB(cfg.DBConn)

	// 3. Собираем слои
	userRepo := repository.NewUserRepository(db)

	// В твоем коде функция называется NewUserService, используем её через пакет services
	authServ := services.NewUserService(userRepo, cfg.JwtSecret)

	authHand := http.NewAuthHandler(authServ)

	r := gin.Default()

	// 4. Роуты (можно прописать прямо тут для простоты)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHand.Register)
		authGroup.POST("/login", authHand.Login)
		authGroup.PUT("/update/:id", authHand.UpdateProfile)
		authGroup.GET("/user/:id", authHand.GetUser)
	}

	// Запуск на порту из конфига
	r.Run(":" + cfg.Port)
}
