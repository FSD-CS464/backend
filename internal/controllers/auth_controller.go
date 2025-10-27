package controllers

import (
	"net/http"

	"fsd-backend/internal/middleware"
	"fsd-backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	Signer *auth.Signer
}

func NewAuthController(s *auth.Signer) *AuthController { return &AuthController{Signer: s} }

// POST /auth/login
func (a *AuthController) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// TODO: replace with real user validation
	userID := req.Username

	access, _ := a.Signer.SignAccess(userID)
	refresh, _ := a.Signer.SignRefresh(userID)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// POST /auth/refresh
func (a *AuthController) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	claims, err := a.Signer.Parse(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	access, _ := a.Signer.SignAccess(claims.UserID)
	c.JSON(http.StatusOK, gin.H{"access_token": access})
}

// GET /auth/me (protected)
func (a *AuthController) Me(c *gin.Context) {
	uid := middleware.UserID(c)
	c.JSON(http.StatusOK, gin.H{"user_id": uid})
}
