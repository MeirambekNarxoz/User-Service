package http

import (
	"context"
	stdhttp "net/http"
	"time"
	"user-service/internal/middleware"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	presence      *services.PresenceService
	jwtSecret     string
	allowedOrigin string
	upgrader      websocket.Upgrader
}

func NewWSHandler(p *services.PresenceService, jwtSecret, allowedOrigin string) *WSHandler {
	h := &WSHandler{
		presence:      p,
		jwtSecret:     jwtSecret,
		allowedOrigin: allowedOrigin,
	}

	h.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *stdhttp.Request) bool {
			origin := r.Header.Get("Origin")
			return allowedOrigin == "*" || origin == allowedOrigin
		},
	}

	return h
}

func (h *WSHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(stdhttp.StatusUnauthorized, gin.H{"error": "token is required"})
		return
	}

	userID, err := middleware.UserIDFromToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(stdhttp.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = h.presence.MarkOnline(context.Background(), userID)

	conn.SetReadLimit(1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return h.presence.MarkOnline(context.Background(), userID)
	})

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
			_ = h.presence.MarkOnline(context.Background(), userID)
		}
	}()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			_ = h.presence.MarkOnline(context.Background(), userID)
		}
	}
}
