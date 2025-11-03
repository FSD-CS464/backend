package controllers

import (
	"context"
	"net/http"

	"fsd-backend/internal/middleware"
	"fsd-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HabitController struct {
	repo *repository.HabitRepo
}

func NewHabitController(db *pgxpool.Pool) *HabitController {
	return &HabitController{repo: repository.NewHabitRepo(db)}
}

// GET /habits - Get all habits for the authenticated user
func (ctl *HabitController) List(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	habits, err := ctl.repo.GetByUserID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": habits})
}

// GET /habits/:id - Get a specific habit by ID
func (ctl *HabitController) GetByID(c *gin.Context) {
	id := c.Param("id")
	h, err := ctl.repo.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}

	// Verify the habit belongs to the authenticated user
	userID := middleware.UserID(c)
	if userID != "" && h.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": h})
}

type createHabitReq struct {
	Title   string `json:"title" binding:"required"`
	Done    bool   `json:"done"`
	Icons   string `json:"icons"`
	Cadence string `json:"cadence" binding:"required"` // "daily" | "everyN-<n_days>" | "weekly-<day_of_the_week>"
}

// POST /habits - Create a new habit for the authenticated user
func (ctl *HabitController) Create(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createHabitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default values if not provided
	done := req.Done
	icons := req.Icons
	if icons == "" {
		icons = "💡"
	}

	h, err := ctl.repo.Create(context.Background(), userID, req.Title, done, icons, req.Cadence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": h})
}

type updateHabitReq struct {
	Title   *string `json:"title"`
	Done    *bool   `json:"done"`
	Icons   *string `json:"icons"`
	Cadence *string `json:"cadence"`
}

// PUT /habits/:id - Update a habit by ID
func (ctl *HabitController) Update(c *gin.Context) {
	id := c.Param("id")
	
	// Verify the habit belongs to the authenticated user
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check if habit exists and belongs to user
	h, err := ctl.repo.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}

	if h.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req updateHabitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the habit
	updatedHabit, err := ctl.repo.Update(c, id, req.Title, req.Done, req.Icons, req.Cadence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedHabit})
}

// DELETE /habits/:id - Delete a habit by ID
func (ctl *HabitController) Delete(c *gin.Context) {
	id := c.Param("id")
	
	// Verify the habit belongs to the authenticated user
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	h, err := ctl.repo.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}

	if h.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := ctl.repo.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

