package middleware

import (
	"errors"
	"strings"

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

func UserIDFromAuthHeader(c *gin.Context, secret string) (uint, error) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return 0, errors.New("authorization header is empty")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return 0, errors.New("invalid auth header")
	}

	return UserIDFromToken(strings.TrimPrefix(auth, prefix), secret)
}
