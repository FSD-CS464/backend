package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct{ db *pgxpool.Pool }
func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	const q = `SELECT id, email, display_name, attrs, created_at, updated_at FROM users WHERE id = $1`
	var u User
	if err := r.db.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) List(ctx context.Context, limit int) ([]User, error) {
	const q = `SELECT id, email, display_name, attrs, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1`
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
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
RETURNING id, email, display_name, attrs, created_at, updated_at`
	var u User
	if err := r.db.QueryRow(ctx, q, email, displayName, attrs).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdateName(ctx context.Context, id, displayName string) (*User, error) {
	const q = `
UPDATE users SET display_name = $2, updated_at = now()
WHERE id = $1
RETURNING id, email, display_name, attrs, created_at, updated_at`
	var u User
	if err := r.db.QueryRow(ctx, q, id, displayName).
		Scan(&u.ID, &u.Email, &u.DisplayName, &u.Attrs, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}
