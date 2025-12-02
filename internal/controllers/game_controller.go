package controllers

import (
	"context"
	"net/http"

	"fsd-backend/internal/middleware"
	"fsd-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameController struct {
	repo *repository.GameRepo
}

func NewGameController(db *pgxpool.Pool) *GameController {
	return &GameController{
		repo: repository.NewGameRepo(db),
	}
}

// GET /game/data - Get user's game data (high scores for all games)
func (g *GameController) GetUserData(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	games, err := g.repo.GetAllByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game data"})
		return
	}

	// Build high scores map
	highScores := make(map[string]int)
	for _, game := range games {
		if score, ok := game.Attrs["high_score"].(float64); ok {
			highScores[game.Title] = int(score)
		} else if score, ok := game.Attrs["high_score"].(int); ok {
			highScores[game.Title] = score
		} else {
			highScores[game.Title] = 0
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"high_scores": highScores,
	})
}

// POST /game/save - Save game data (e.g., high score)
func (g *GameController) SaveGameData(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	var req struct {
		GameType string `json:"game_type" binding:"required"`
		Score    int    `json:"score" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Upsert high score (only updates if new score is higher)
	game, err := g.repo.UpsertHighScore(context.Background(), userID, req.GameType, req.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save game data"})
		return
	}

	// Extract high score from response
	highScore := 0
	if score, ok := game.Attrs["high_score"].(float64); ok {
		highScore = int(score)
	} else if score, ok := game.Attrs["high_score"].(int); ok {
		highScore = score
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Game data saved successfully",
		"user_id":    userID,
		"game_type":  req.GameType,
		"score":      req.Score,
		"high_score": highScore,
	})
}
