package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GithubAuthHandler struct {
	authService *services.AuthService
	oauthConf   *oauth2.Config
}

func NewGithubAuthHandler(authService *services.AuthService) *GithubAuthHandler {
	return &GithubAuthHandler{
		authService: authService,
		oauthConf: &oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  "http://localhost:8080/api/auth/github/callback",
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
		},
	}
}

func (h *GithubAuthHandler) Login(c *gin.Context) {
	url := h.oauthConf.AuthCodeURL("state-github")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *GithubAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	token, err := h.oauthConf.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange Github token"})
		return
	}

	client := h.oauthConf.Client(context.Background(), token)
	
	// 1. Get basic user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get Github user"})
		return
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	json.NewDecoder(resp.Body).Decode(&githubUser)

	// 2. GitHub might return empty email if it's private. Fetch emails separately.
	if githubUser.Email == "" {
		respEmails, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer respEmails.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			json.NewDecoder(respEmails.Body).Decode(&emails)
			for _, e := range emails {
				if e.Primary {
					githubUser.Email = e.Email
					break
				}
			}
		}
	}

	// 3. Authenticate in our system
	jwtToken, err := h.authService.LoginWithGithub(
		fmt.Sprintf("%d", githubUser.ID),
		githubUser.Email,
		githubUser.Name,
		githubUser.Login,
		githubUser.AvatarURL,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Redirect back to frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	c.Redirect(http.StatusPermanentRedirect, frontendURL+"/login-success?token="+jwtToken)
}
