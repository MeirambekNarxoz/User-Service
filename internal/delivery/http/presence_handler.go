package http

import (
	stdhttp "net/http"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
)

type PresenceHandler struct {
	presence *services.PresenceService
}

type UsersStatusRequest struct {
	UserIDs []uint `json:"user_ids"`
}

func NewPresenceHandler(p *services.PresenceService) *PresenceHandler {
	return &PresenceHandler{presence: p}
}

func (h *PresenceHandler) GetOnlineCount(c *gin.Context) {
	count, err := h.presence.GetOnlineCount(c.Request.Context())
	if err != nil {
		c.JSON(stdhttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(stdhttp.StatusOK, gin.H{
		"online_count": count,
	})
}

func (h *PresenceHandler) GetUsersStatus(c *gin.Context) {
	var req UsersStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": "неверный список user_ids"})
		return
	}

	users, err := h.presence.GetUsersStatus(c.Request.Context(), req.UserIDs)
	if err != nil {
		c.JSON(stdhttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	onlineCount := 0
	for _, u := range users {
		if u.Online {
			onlineCount++
		}
	}

	c.JSON(stdhttp.StatusOK, gin.H{
		"users":        users,
		"online_count": onlineCount,
	})
}
