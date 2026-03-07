package routes

import (
	delivery "user-service/internal/delivery/http"
	"user-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authHandler *delivery.AuthHandler,
	wsHandler *delivery.WSHandler,
	presenceHandler *delivery.PresenceHandler,
	internalAPIToken string,
) {
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/users/:id", authHandler.GetUser)
		authGroup.PUT("/update/:id", authHandler.UpdateProfile)
	}

	r.GET("/api/presence/ws", wsHandler.Connect)

	internal := r.Group("/internal")
	internal.Use(middleware.InternalOnlyMiddleware(internalAPIToken))
	{
		internal.GET("/presence/count", presenceHandler.GetOnlineCount)
		internal.POST("/presence/users", presenceHandler.GetUsersStatus)
	}
}
