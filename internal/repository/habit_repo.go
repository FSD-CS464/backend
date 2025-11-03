package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HabitRepo struct{ db *pgxpool.Pool }

func NewHabitRepo(db *pgxpool.Pool) *HabitRepo {
	return &HabitRepo{db: db}
}

func (r *HabitRepo) GetByID(ctx context.Context, id string) (*Habit, error) {
	const q = `SELECT id, user_id, title, done, icons, cadence, created_at, updated_at FROM habits WHERE id = $1`
	var h Habit
	if err := r.db.QueryRow(ctx, q, id).
		Scan(&h.ID, &h.UserID, &h.Title, &h.Done, &h.Icons, &h.Cadence, &h.CreatedAt, &h.UpdatedAt); err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HabitRepo) GetByUserID(ctx context.Context, userID string) ([]Habit, error) {
	const q = `SELECT id, user_id, title, done, icons, cadence, created_at, updated_at 
	           FROM habits WHERE user_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Habit
	for rows.Next() {
		var h Habit
		if err := rows.Scan(&h.ID, &h.UserID, &h.Title, &h.Done, &h.Icons, &h.Cadence, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *HabitRepo) Create(ctx context.Context, userID, title string, done bool, icons, cadence string) (*Habit, error) {
	const q = `
INSERT INTO habits (user_id, title, done, icons, cadence)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, title, done, icons, cadence, created_at, updated_at`
	var h Habit
	if err := r.db.QueryRow(ctx, q, userID, title, done, icons, cadence).
		Scan(&h.ID, &h.UserID, &h.Title, &h.Done, &h.Icons, &h.Cadence, &h.CreatedAt, &h.UpdatedAt); err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HabitRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM habits WHERE id = $1`, id)
	return err
}

