package controllers

import (
	"context"
	"net/http"
	"strings"

	"fsd-backend/internal/auth"
	"fsd-backend/internal/middleware"
	"fsd-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthController struct {
	Signer *auth.Signer
	Users  *repository.UserRepo
}

func NewAuthController(s *auth.Signer, db *pgxpool.Pool) *AuthController {
	return &AuthController{
		Signer: s,
		Users:  repository.NewUserRepo(db),
	}
}

// POST /auth/register
func (a *AuthController) Register(c *gin.Context) {
	var req struct {
		Email       string                 `json:"email" binding:"required,email"`
		DisplayName string                 `json:"display_name" binding:"required"`
		Password    string                 `json:"password" binding:"required,min=4"`
		Attrs       map[string]any         `json:"attrs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pwHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	u, err := a.Users.CreateWithPassword(c, strings.ToLower(req.Email), req.DisplayName, pwHash, req.Attrs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
		return
	}

	access, _ := a.Signer.SignAccess(u.ID)
	refresh, _ := a.Signer.SignRefresh(u.ID)

	c.JSON(http.StatusCreated, gin.H{
		"user":          gin.H{"id": u.ID, "email": u.Email, "display_name": u.DisplayName},
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// POST /auth/login
func (a *AuthController) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	identifier := req.Email
	if identifier == "" {
		identifier = req.Username
	}
	identifier = strings.ToLower(identifier)

	u, err := a.Users.GetByEmail(c, identifier)
	if err != nil || u.PasswordHash == "" || !auth.CheckPassword(u.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	access, _ := a.Signer.SignAccess(u.ID)
	refresh, _ := a.Signer.SignRefresh(u.ID)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"user": gin.H{
			"id":            u.ID,
			"email":         u.Email,
			"display_name":  u.DisplayName,
		},
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

// GET /auth/me
func (a *AuthController) Me(c *gin.Context) {
	uid := middleware.UserID(c)
	u, err := a.Users.GetByID(context.Background(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":            u.ID,
		"email":         u.Email,
		"display_name":  u.DisplayName,
		"attrs":         u.Attrs,
	})
}
