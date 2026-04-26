package main

import (
	"log"
	"user-service/internal/config"
	"user-service/internal/database"
	delivery "user-service/internal/delivery/http"
	"user-service/internal/models"
	"user-service/internal/rabbitmq"
	"user-service/internal/repository"
	"user-service/internal/routes"
	"user-service/internal/services"
	"user-service/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db := database.InitDB(cfg.DBConn)
	db.AutoMigrate(&models.User{})
	rdb := database.InitRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	userRepo := repository.NewUserRepository(db)

	// Init MinIO
	minioClient := storage.NewMinioClient(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOUseSSL)

	// Init RabbitMQ Producer
	rabbitProducer, err := rabbitmq.NewRabbitMQProducer(cfg.RabbitURL)
	if err != nil {
		log.Printf("Warning: failed to init RabbitMQ: %v", err)
	} else {
		defer rabbitProducer.Close()
	}

	// Init Email Service (SMTP)
	emailService := services.NewEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPass,
		cfg.SMTPFromName,
	)

	authService := services.NewUserService(userRepo, cfg.JwtSecret, rdb, rabbitProducer, emailService)
	authHandler := delivery.NewAuthHandler(authService, minioClient)
	googleAuthHandler := delivery.NewGoogleAuthHandler(authService)
	githubAuthHandler := delivery.NewGithubAuthHandler(authService)
	linkedinAuthHandler := delivery.NewLinkedinAuthHandler(authService)

	presenceService := services.NewPresenceService(rdb)
	wsHandler := delivery.NewWSHandler(presenceService, cfg.JwtSecret, cfg.WSAllowedOrigin)
	presenceHandler := delivery.NewPresenceHandler(presenceService)

	r := gin.Default()
	routes.SetupRoutes(
		r,
		authHandler,
		googleAuthHandler,
		githubAuthHandler,
		linkedinAuthHandler,
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
