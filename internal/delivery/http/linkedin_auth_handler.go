package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

type LinkedinAuthHandler struct {
	authService *services.AuthService
	oauthConf   *oauth2.Config
}

func NewLinkedinAuthHandler(authService *services.AuthService) *LinkedinAuthHandler {
	return &LinkedinAuthHandler{
		authService: authService,
		oauthConf: &oauth2.Config{
			ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
			ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
			RedirectURL:  "http://localhost:8080/api/auth/linkedin/callback",
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     linkedin.Endpoint,
		},
	}
}

func (h *LinkedinAuthHandler) Login(c *gin.Context) {
	url := h.oauthConf.AuthCodeURL("state-linkedin")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *LinkedinAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	token, err := h.oauthConf.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange Linkedin token"})
		return
	}

	client := h.oauthConf.Client(context.Background(), token)
	
	// LinkedIn uses OpenID Connect userinfo endpoint
	resp, err := client.Get("https://api.linkedin.com/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get Linkedin user info"})
		return
	}
	defer resp.Body.Close()

	var linkedinUser struct {
		Sub           string `json:"sub"` // Unique ID
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Email         string `json:"email"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&linkedinUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode Linkedin user"})
		return
	}

	jwtToken, err := h.authService.LoginWithLinkedin(
		linkedinUser.Sub,
		linkedinUser.Email,
		linkedinUser.GivenName,
		linkedinUser.FamilyName,
		linkedinUser.Picture,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	c.Redirect(http.StatusPermanentRedirect, frontendURL+"/login-success?token="+jwtToken)
}
