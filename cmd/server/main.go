package main

import (
	"log"
	"user-service/internal/config"
	"user-service/internal/database"
	delivery "user-service/internal/delivery/http"
	"user-service/internal/repository"
	"user-service/internal/routes"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db := database.InitDB(cfg.DBConn)
	rdb := database.InitRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	userRepo := repository.NewUserRepository(db)
	authService := services.NewUserService(userRepo, cfg.JwtSecret)
	authHandler := delivery.NewAuthHandler(authService)

	presenceService := services.NewPresenceService(rdb)
	wsHandler := delivery.NewWSHandler(presenceService, cfg.JwtSecret, cfg.WSAllowedOrigin)
	presenceHandler := delivery.NewPresenceHandler(presenceService)

	r := gin.Default()
	
	routes.SetupRoutes(
		r,
		authHandler,
		wsHandler,
		presenceHandler,
		cfg.InternalAPIToken,
		cfg.JwtSecret,
	)

	log.Println("Server started on port:", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
