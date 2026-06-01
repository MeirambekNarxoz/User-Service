package middleware

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"user-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func UserIDFromToken(tokenString, secret string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	id, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("user_id not found in token")
	}

	return uint(id), nil
}

// GetUserID retrieves the UserID safely from the context
func GetUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := val.(uint)
	return userID, ok
}

// AuthMiddleware
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-Id")

		if userIDStr == "" {
			log.Printf("DEBUG: X-User-Id header is missing! All headers: %v", c.Request.Header)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "x-user-id header is missing"})
			return
		}
		userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "x-user-id header is invalid"})
			return
		}
		rolesStr := c.GetHeader("X-User-Roles")
		var role models.Role
		if strings.Contains(rolesStr, "ROLE_ADMIN") {
			role = models.RoleAdmin
		} else if strings.Contains(rolesStr, "ROLE_MODERATOR") {
			role = models.RoleModerator
		} else {
			role = models.RoleUser
		}
		c.Set("user_id", uint(userIDUint))
		c.Set("role", role)
		c.Next()
	}
}
