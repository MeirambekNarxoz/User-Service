package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuthHandler struct {
	authService *services.AuthService
	oauthConf   *oauth2.Config
}

func NewGoogleAuthHandler(authService *services.AuthService) *GoogleAuthHandler {
	return &GoogleAuthHandler{
		authService: authService,
		oauthConf: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  "http://localhost:8080/api/auth/google/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (h *GoogleAuthHandler) Login(c *gin.Context) {
	url := h.oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *GoogleAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	token, err := h.oauthConf.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := h.oauthConf.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
		Picture   string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode user info"})
		return
	}

	jwtToken, err := h.authService.LoginWithGoogle(
		googleUser.ID,
		googleUser.Email,
		googleUser.FirstName,
		googleUser.LastName,
		googleUser.Picture,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redirect back to frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // Default Vite port
	}
	redirectTarget := frontendURL + "/login-success?token=" + jwtToken
	c.Redirect(http.StatusPermanentRedirect, redirectTarget)
}
