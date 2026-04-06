package middleware

import (
	"errors"
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

// AuthMiddleware extracts X-User-Id and X-User-Roles from headers (injected by API Gateway)
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-Id")
		
		if userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "x-user-id header is missing"})
			return
		}

		userIDFloat, err := strconv.ParseFloat(userIDStr, 64)
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

		c.Set("user_id", uint(userIDFloat))
		c.Set("role", role)
		c.Next()
	}
}
