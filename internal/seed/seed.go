package seed

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User  struct{ Email, DisplayName string; Attrs map[string]any }
type Pet   struct{ UserEmail, Name, Species string; Attrs map[string]any }
type Habit struct{ UserEmail, Title, Cadence string; Attrs map[string]any }
type Game  struct{ UserEmail, Title, Status string; Attrs map[string]any }

type Data struct {
	Users  []User
	Pets   []Pet
	Habits []Habit
	Games  []Game
}

func Run(ctx context.Context, pool *pgxpool.Pool, d Data) error {
	// Upsert users by email
	for _, u := range d.Users {
		attr, _ := json.Marshal(u.Attrs)
		_, err := pool.Exec(ctx, `
INSERT INTO users (id, email, display_name, attrs)
VALUES (gen_random_uuid(), $1, $2, COALESCE($3, '{}'::JSONB))
ON CONFLICT (email) DO UPDATE
  SET display_name = EXCLUDED.display_name,
      attrs = EXCLUDED.attrs,
      updated_at = now()`, u.Email, u.DisplayName, attr)
		if err != nil { return err }
	}

	getUID := func(email string) (string, error) {
		var id string
		err := pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
		return id, err
	}

	// Pets: upsert by (user_id, name)
	for _, p := range d.Pets {
		uid, err := getUID(p.UserEmail); if err != nil { return err }
		attr, _ := json.Marshal(p.Attrs)
		_, err = pool.Exec(ctx, `
INSERT INTO pets (id, user_id, name, species, attrs)
VALUES (gen_random_uuid(), $1, $2, $3, COALESCE($4, '{}'::JSONB))
ON CONFLICT (user_id, name) DO UPDATE
  SET species = EXCLUDED.species, attrs = EXCLUDED.attrs, updated_at = now()`,
			uid, p.Name, p.Species, attr)
		if err != nil { return err }
	}

	// Habits: upsert by (user_id, title)
	for _, h := range d.Habits {
		uid, err := getUID(h.UserEmail); if err != nil { return err }
		attr, _ := json.Marshal(h.Attrs)
		_, err = pool.Exec(ctx, `
INSERT INTO habits (id, user_id, title, cadence, attrs)
VALUES (gen_random_uuid(), $1, $2, $3, COALESCE($4, '{}'::JSONB))
ON CONFLICT (user_id, title) DO UPDATE
  SET cadence = EXCLUDED.cadence, attrs = EXCLUDED.attrs, updated_at = now()`,
			uid, h.Title, h.Cadence, attr)
		if err != nil { return err }
	}

	// Games: upsert by (user_id, title)
	for _, g := range d.Games {
		uid, err := getUID(g.UserEmail); if err != nil { return err }
		attr, _ := json.Marshal(g.Attrs)
		_, err = pool.Exec(ctx, `
INSERT INTO games (id, user_id, title, status, attrs)
VALUES (gen_random_uuid(), $1, $2, $3, COALESCE($4, '{}'::JSONB))
ON CONFLICT (user_id, title) DO UPDATE
  SET status = EXCLUDED.status, attrs = EXCLUDED.attrs, updated_at = now()`,
			uid, g.Title, g.Status, attr)
		if err != nil { return err }
	}

	return nil
}
