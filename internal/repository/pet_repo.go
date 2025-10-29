package repository

import (
	"context"

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
