package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	const q = `SELECT id, email, display_name, password_hash, attrs, created_at, updated_at FROM users WHERE id = $1`
	var u User
	if err := r.db.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	const q = `SELECT id, email, display_name, password_hash, attrs, created_at, updated_at FROM users WHERE email = $1`
	var u User
	if err := r.db.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) List(ctx context.Context, limit int) ([]User, error) {
	const q = `SELECT id, email, display_name, password_hash, attrs, created_at, updated_at
	           FROM users ORDER BY created_at DESC LIMIT $1`
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserRepo) Create(ctx context.Context, email, displayName string, attrs map[string]any) (*User, error) {
	const q = `
INSERT INTO users (email, display_name, attrs)
VALUES ($1, $2, COALESCE($3, '{}'::JSONB))
RETURNING id, email, display_name, password_hash, attrs, created_at, updated_at`
	var u User
	if err := r.db.QueryRow(ctx, q, email, displayName, attrs).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) CreateWithPassword(ctx context.Context, email, displayName, passwordHash string, attrs map[string]any) (*User, error) {
	const q = `
INSERT INTO users (email, display_name, password_hash, attrs)
VALUES ($1, $2, $3, COALESCE($4, '{}'::JSONB))
RETURNING id, email, display_name, password_hash, attrs, created_at, updated_at`
	var u User
	if err := r.db.QueryRow(ctx, q, email, displayName, passwordHash, attrs).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdateName(ctx context.Context, id, displayName string) (*User, error) {
	const q = `
UPDATE users SET display_name = $2, updated_at = now()
WHERE id = $1
RETURNING id, email, display_name, password_hash, attrs, created_at, updated_at`
	var u User
	if err := r.db.QueryRow(ctx, q, id, displayName).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// GetEnergy gets the user's energy from attrs, defaulting to 30 if not set
func (r *UserRepo) GetEnergy(ctx context.Context, userID string) (int, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	
	if energy, ok := user.Attrs["energy"].(float64); ok {
		return int(energy), nil
	} else if energy, ok := user.Attrs["energy"].(int); ok {
		return energy, nil
	}
	// Default energy is 30
	return 30, nil
}

// UpdateEnergy updates the user's energy in attrs
func (r *UserRepo) UpdateEnergy(ctx context.Context, userID string, energy int) error {
	// Ensure energy is within bounds
	if energy < 0 {
		energy = 0
	} else if energy > 100 {
		energy = 100
	}
	
	// Get current user to preserve other attrs
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	
	// Update attrs
	if user.Attrs == nil {
		user.Attrs = make(map[string]any)
	}
	user.Attrs["energy"] = energy
	
	// Update in database
	attrsJSON, err := json.Marshal(user.Attrs)
	if err != nil {
		return err
	}
	
	const q = `UPDATE users SET attrs = $2, updated_at = now() WHERE id = $1`
	_, err = r.db.Exec(ctx, q, userID, attrsJSON)
	return err
}