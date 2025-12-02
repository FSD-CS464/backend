package controllers

import (
	"context"
	"log"
	"net/http"

	"fsd-backend/internal/middleware"
	"fsd-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameController struct {
	repo     *repository.GameRepo
	userRepo *repository.UserRepo
	petRepo  *repository.PetRepo
}

func NewGameController(db *pgxpool.Pool) *GameController {
	return &GameController{
		repo:     repository.NewGameRepo(db),
		userRepo: repository.NewUserRepo(db),
		petRepo:  repository.NewPetRepo(db),
	}
}

// GET /game/data - Get user's game data (high scores for all games, energy)
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

	// Get energy
	energy, err := g.userRepo.GetEnergy(c.Request.Context(), userID)
	if err != nil {
		energy = 30 // Default if error
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"high_scores": highScores,
		"energy":      energy,
	})
}

// POST /game/save - Save game data (e.g., high score) and increase mood
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

	// Increase mood FIRST - this happens for EVERY game, regardless of high score
	// Jump Rope: adds score to mood
	// Sunny Says: adds 2x score to mood
	moodIncrease := req.Score
	if req.GameType == "Sunny Says" {
		moodIncrease = req.Score * 2
	}

	if err := g.petRepo.AddMood(c.Request.Context(), userID, moodIncrease); err != nil {
		// Log error but don't fail the request
		// Mood update failure shouldn't prevent score saving
		log.Printf("ERROR: Failed to increase mood for user %s: %v", userID, err)
	}

	// Upsert high score (only updates if new score is higher)
	// This is separate from mood increase - mood increases for every game
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

// POST /game/check-energy - Check if user has enough energy to play a game
func (g *GameController) CheckEnergy(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	var req struct {
		GameType string `json:"game_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current energy
	energy, err := g.userRepo.GetEnergy(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch energy"})
		return
	}

	// Determine energy cost
	var requiredEnergy int
	switch req.GameType {
	case "Jump Rope":
		requiredEnergy = 10
	case "Sunny Says":
		requiredEnergy = 15
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game type"})
		return
	}

	hasEnough := energy >= requiredEnergy

	c.JSON(http.StatusOK, gin.H{
		"has_enough_energy": hasEnough,
		"current_energy":    energy,
		"required_energy":   requiredEnergy,
	})
}

// POST /game/deduct-energy - Deduct energy before starting a game
func (g *GameController) DeductEnergy(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	var req struct {
		GameType string `json:"game_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current energy
	currentEnergy, err := g.userRepo.GetEnergy(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch energy"})
		return
	}

	// Determine energy cost
	var energyCost int
	switch req.GameType {
	case "Jump Rope":
		energyCost = 10
	case "Sunny Says":
		energyCost = 15
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game type"})
		return
	}

	// Check if user has enough energy
	if currentEnergy < energyCost {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "Insufficient energy",
			"current_energy":  currentEnergy,
			"required_energy": energyCost,
		})
		return
	}

	// Deduct energy
	newEnergy := currentEnergy - energyCost
	if err := g.userRepo.UpdateEnergy(c.Request.Context(), userID, newEnergy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update energy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Energy deducted successfully",
		"previous_energy": currentEnergy,
		"new_energy":      newEnergy,
		"energy_cost":     energyCost,
	})
}

// GET /game/mood - Get pet's mood (with passive drain calculation)
func (g *GameController) GetMood(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	mood, err := g.petRepo.GetMood(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch mood"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mood": mood})
}
