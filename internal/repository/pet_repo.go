package repository

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PetRepo struct{ db *pgxpool.Pool }

func NewPetRepo(db *pgxpool.Pool) *PetRepo { return &PetRepo{db: db} }

func (r *PetRepo) GetByID(ctx context.Context, id string) (*Pet, error) {
	const q = `SELECT id, user_id, name, species, attrs, created_at, updated_at FROM pets WHERE id = $1`
	var p Pet
	if err := r.db.QueryRow(ctx, q, id).
		Scan(&p.ID, &p.UserID, &p.Name, &p.Species, &p.Attrs, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PetRepo) Create(ctx context.Context, userID, name, species string, attrs map[string]any) (*Pet, error) {
	const q = `
INSERT INTO pets (user_id, name, species, attrs)
VALUES ($1, $2, $3, COALESCE($4, '{}'::JSONB))
RETURNING id, user_id, name, species, attrs, created_at, updated_at`
	var p Pet
	if err := r.db.QueryRow(ctx, q, userID, name, species, attrs).
		Scan(&p.ID, &p.UserID, &p.Name, &p.Species, &p.Attrs, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByUserID gets the first pet for a user (assuming one pet per user)
func (r *PetRepo) GetByUserID(ctx context.Context, userID string) (*Pet, error) {
	const q = `SELECT id, user_id, name, species, attrs, created_at, updated_at FROM pets WHERE user_id = $1 LIMIT 1`
	var p Pet
	var attrsJSON []byte
	if err := r.db.QueryRow(ctx, q, userID).
		Scan(&p.ID, &p.UserID, &p.Name, &p.Species, &attrsJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(attrsJSON, &p.Attrs); err != nil {
		p.Attrs = make(map[string]any)
	}
	return &p, nil
}

// GetMood gets the pet's mood with passive drain calculation
// Mood drains -1 every 25 minutes since last update
func (r *PetRepo) GetMood(ctx context.Context, userID string) (int, error) {
	pet, err := r.GetByUserID(ctx, userID)
	if err != nil {
		// If pet doesn't exist, return default mood
		return 50, nil
	}

	// Get current mood from attrs
	var currentMood int
	if mood, ok := pet.Attrs["mood"].(float64); ok {
		currentMood = int(mood)
	} else if mood, ok := pet.Attrs["mood"].(int); ok {
		currentMood = mood
	} else {
		// Default mood is 50
		currentMood = 50
	}

	// Get last mood update time
	var lastUpdate time.Time
	if lastUpdateStr, ok := pet.Attrs["mood_last_updated"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, lastUpdateStr); err == nil {
			lastUpdate = parsed
		} else {
			lastUpdate = pet.UpdatedAt
		}
	} else {
		lastUpdate = pet.UpdatedAt
	}

	// Calculate passive drain: -1 every 25 minutes
	minutesSinceUpdate := time.Since(lastUpdate).Minutes()
	drainAmount := int(math.Floor(minutesSinceUpdate / 25.0))

	// Apply drain (mood can't go below 0)
	newMood := currentMood - drainAmount
	if newMood < 0 {
		newMood = 0
	}

	// If mood changed, update it in the database
	if drainAmount > 0 {
		if err := r.UpdateMood(ctx, userID, newMood); err != nil {
			// If update fails, return the calculated mood anyway
			return newMood, nil
		}
	}

	return newMood, nil
}

// UpdateMood updates the pet's mood and records the update time
// If pet doesn't exist, creates a default pet with the specified mood
func (r *PetRepo) UpdateMood(ctx context.Context, userID string, mood int) error {
	// Ensure mood is within bounds
	if mood < 0 {
		mood = 0
	} else if mood > 100 {
		mood = 100
	}

	// Get current pet to preserve other attrs
	pet, err := r.GetByUserID(ctx, userID)
	if err != nil {
		// If pet doesn't exist, create a default pet
		attrs := make(map[string]any)
		attrs["mood"] = mood
		attrs["mood_last_updated"] = time.Now().Format(time.RFC3339)

		// Try to create pet
		_, createErr := r.Create(ctx, userID, "My Pet", "default", attrs)
		if createErr != nil {
			// If creation failed (e.g., pet was created concurrently), try to get it again
			pet, err = r.GetByUserID(ctx, userID)
			if err != nil {
				// Still can't get pet, return original create error
				return createErr
			}
			// Pet was created by another request, continue with update below
		} else {
			// Pet created successfully, we're done
			return nil
		}
	}

	// Update attrs (either pet existed or was just created)
	if pet.Attrs == nil {
		pet.Attrs = make(map[string]any)
	}
	pet.Attrs["mood"] = mood
	pet.Attrs["mood_last_updated"] = time.Now().Format(time.RFC3339)

	// Update in database
	attrsJSON, err := json.Marshal(pet.Attrs)
	if err != nil {
		return err
	}

	const q = `UPDATE pets SET attrs = $2, updated_at = now() WHERE user_id = $1`
	_, err = r.db.Exec(ctx, q, userID, attrsJSON)
	return err
}

// AddMood adds to the pet's mood (for games)
// This function gets the current mood (with passive drain applied), adds the amount, and updates
func (r *PetRepo) AddMood(ctx context.Context, userID string, amount int) error {
	// Get current mood (this will apply passive drain if needed)
	currentMood, err := r.GetMood(ctx, userID)
	if err != nil {
		return err
	}

	// Add the mood increase
	newMood := currentMood + amount
	// UpdateMood will clamp to 0-100 and update the timestamp
	return r.UpdateMood(ctx, userID, newMood)
}
