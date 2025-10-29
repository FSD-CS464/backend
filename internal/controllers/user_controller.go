package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"fsd-backend/internal/repository"
)

type UserController struct{ repo *repository.UserRepo }
func NewUserController(db *pgxpool.Pool) *UserController { return &UserController{repo: repository.NewUserRepo(db)} }

func (ctl *UserController) List(c *gin.Context) {
	limit := 50
	if s := c.Query("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= 200 { limit = v }
	}
	users, err := ctl.repo.List(c, limit)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (ctl *UserController) GetByID(c *gin.Context) {
	id := c.Param("id")
	u, err := ctl.repo.GetByID(c, id)
	if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "user not found"}); return }
	c.JSON(http.StatusOK, gin.H{"data": u})
}

type createUserReq struct {
	Email       string                 `json:"email" binding:"required,email"`
	DisplayName string                 `json:"display_name" binding:"required"`
	Attrs       map[string]any         `json:"attrs"`
}

func (ctl *UserController) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}
	u, err := ctl.repo.Create(context.Background(), req.Email, req.DisplayName, req.Attrs)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusCreated, gin.H{"data": u})
}

func (ctl *UserController) UpdateName(c *gin.Context) {
	id := c.Param("id")
	var body struct{ DisplayName string `json:"display_name" binding:"required"` }
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
	}
	u, err := ctl.repo.UpdateName(c, id, body.DisplayName)
	if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "user not found"}); return }
	c.JSON(http.StatusOK, gin.H{"data": u})
}

func (ctl *UserController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctl.repo.Delete(c, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"}); return
	}
	c.Status(http.StatusNoContent)
}
