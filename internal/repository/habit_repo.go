package repository

import (
	"context"
	"fmt"
	"strings"

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

func (r *HabitRepo) Update(ctx context.Context, id string, title *string, done *bool, icons *string, cadence *string) (*Habit, error) {
	// Build dynamic UPDATE query based on provided fields
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argPos))
		args = append(args, *title)
		argPos++
	}
	if done != nil {
		updates = append(updates, fmt.Sprintf("done = $%d", argPos))
		args = append(args, *done)
		argPos++
	}
	if icons != nil {
		updates = append(updates, fmt.Sprintf("icons = $%d", argPos))
		args = append(args, *icons)
		argPos++
	}
	if cadence != nil {
		updates = append(updates, fmt.Sprintf("cadence = $%d", argPos))
		args = append(args, *cadence)
		argPos++
	}

	if len(updates) == 0 {
		// No fields to update, just return the existing habit
		return r.GetByID(ctx, id)
	}

	// Add updated_at
	updates = append(updates, "updated_at = NOW()")
	
	// Add id to args for WHERE clause
	args = append(args, id)

	// Build the query
	q := fmt.Sprintf(`UPDATE habits SET %s WHERE id = $%d
		RETURNING id, user_id, title, done, icons, cadence, created_at, updated_at`,
		strings.Join(updates, ", "), argPos)

	var h Habit
	if err := r.db.QueryRow(ctx, q, args...).
		Scan(&h.ID, &h.UserID, &h.Title, &h.Done, &h.Icons, &h.Cadence, &h.CreatedAt, &h.UpdatedAt); err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HabitRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM habits WHERE id = $1`, id)
	return err
}

