package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepo struct{ db *pgxpool.Pool }

func NewGameRepo(db *pgxpool.Pool) *GameRepo {
	return &GameRepo{db: db}
}

func (r *GameRepo) GetByUserIDAndTitle(ctx context.Context, userID, title string) (*Game, error) {
	const q = `SELECT id, user_id, title, status, attrs, created_at, updated_at 
	           FROM games WHERE user_id = $1 AND title = $2`
	var g Game
	var attrsJSON []byte
	if err := r.db.QueryRow(ctx, q, userID, title).
		Scan(&g.ID, &g.UserID, &g.Title, &g.Status, &attrsJSON, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(attrsJSON, &g.Attrs); err != nil {
		g.Attrs = make(map[string]any)
	}
	return &g, nil
}

func (r *GameRepo) GetAllByUserID(ctx context.Context, userID string) ([]Game, error) {
	const q = `SELECT id, user_id, title, status, attrs, created_at, updated_at 
	           FROM games WHERE user_id = $1 ORDER BY title ASC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var attrsJSON []byte
		if err := rows.Scan(&g.ID, &g.UserID, &g.Title, &g.Status, &attrsJSON, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(attrsJSON, &g.Attrs); err != nil {
			g.Attrs = make(map[string]any)
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

func (r *GameRepo) UpsertHighScore(ctx context.Context, userID, title string, highScore int) (*Game, error) {
	// First, try to get existing game
	existing, err := r.GetByUserIDAndTitle(ctx, userID, title)

	var currentHighScore int
	if err == nil && existing != nil {
		// Extract current high score from attrs
		if score, ok := existing.Attrs["high_score"].(float64); ok {
			currentHighScore = int(score)
		} else if score, ok := existing.Attrs["high_score"].(int); ok {
			currentHighScore = score
		}

		// Only update if new score is higher
		if highScore <= currentHighScore {
			return existing, nil
		}
	}

	// Prepare attrs with high_score
	attrs := make(map[string]any)
	if existing != nil {
		// Preserve existing attrs
		for k, v := range existing.Attrs {
			attrs[k] = v
		}
	}
	attrs["high_score"] = highScore

	attrsJSON, err := json.Marshal(attrs)
	if err != nil {
		return nil, err
	}

	// Upsert: insert or update
	const q = `
INSERT INTO games (user_id, title, status, attrs)
VALUES ($1, $2, 'active', $3)
ON CONFLICT (user_id, title) DO UPDATE
  SET attrs = EXCLUDED.attrs,
      updated_at = now()
RETURNING id, user_id, title, status, attrs, created_at, updated_at`

	var g Game
	var attrsJSONResult []byte
	if err := r.db.QueryRow(ctx, q, userID, title, attrsJSON).
		Scan(&g.ID, &g.UserID, &g.Title, &g.Status, &attrsJSONResult, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(attrsJSONResult, &g.Attrs); err != nil {
		g.Attrs = attrs
	}

	return &g, nil
}

func (r *GameRepo) GetHighScore(ctx context.Context, userID, title string) (int, error) {
	game, err := r.GetByUserIDAndTitle(ctx, userID, title)
	if err != nil {
		// Game doesn't exist, return 0
		return 0, nil
	}

	if score, ok := game.Attrs["high_score"].(float64); ok {
		return int(score), nil
	}
	if score, ok := game.Attrs["high_score"].(int); ok {
		return score, nil
	}

	return 0, nil
}
