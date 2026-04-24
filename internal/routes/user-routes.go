package routes

import (
	delivery "user-service/internal/delivery/http"
	"user-service/internal/middleware"
	"user-service/internal/models"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authHandler *delivery.AuthHandler,
	wsHandler *delivery.WSHandler,
	presenceHandler *delivery.PresenceHandler,
	internalAPIToken string,
	jwtSecret string,
) {
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register/send-code", authHandler.RegisterSendCode)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/password/forgot", authHandler.ForgotPasswordSendCode)
		authGroup.POST("/password/reset", authHandler.ResetPassword)
	}

	usersGroup := r.Group("/api/users")
	usersGroup.Use(middleware.AuthMiddleware(jwtSecret))
	{
		usersGroup.GET("/:id", authHandler.GetUser)
		usersGroup.GET("/search", authHandler.SearchUsers)
		usersGroup.PUT("/update/:id", authHandler.UpdateProfile)
		usersGroup.POST("/avatar", authHandler.UploadAvatar)

	}

	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(jwtSecret), middleware.RoleMiddleware(models.RoleAdmin))
	{
		adminGroup.GET("/users", authHandler.GetAllUsers)
		adminGroup.PUT("/users/:id/role", authHandler.UpdateRole)
	}

	// Moderation routes (both ADMIN and MODERATOR can block/unblock users)
	moderationGroup := r.Group("/api/admin")
	moderationGroup.Use(middleware.AuthMiddleware(jwtSecret), middleware.RoleMiddleware(models.RoleAdmin, models.RoleModerator))
	{
		moderationGroup.POST("/users/:id/block", authHandler.BlockUser)
		moderationGroup.POST("/users/:id/unblock", authHandler.UnblockUser)
	}

	r.GET("/api/presence/ws", wsHandler.Connect)

	internal := r.Group("/internal")
	internal.Use(middleware.InternalOnlyMiddleware(internalAPIToken))
	{
		internal.GET("/presence/count", presenceHandler.GetOnlineCount)
		internal.POST("/presence/users", presenceHandler.GetUsersStatus)
	}
}
