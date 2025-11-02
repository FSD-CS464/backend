package controllers

import (
	"net/http"

	"fsd-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type GameController struct{}

func NewGameController() *GameController {
	return &GameController{}
}

// GET /game/data - Get user's game data
func (g *GameController) GetUserData(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	// TODO: Query CockroachDB for user's game data (highscores, pet age/emotion)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "Game data retrieved successfully",
		// Add game data fields
	})
}

// POST /game/save - Save game data (e.g., score, progress)
func (g *GameController) SaveGameData(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	var req struct {
		GameType string                 `json:"game_type" binding:"required"`
		Score    int                    `json:"score"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Save to CockroachDB
	// For now it just returns success response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Game data saved successfully",
		"user_id":   userID,
		"game_type": req.GameType,
		"score":     req.Score,
	})
}
