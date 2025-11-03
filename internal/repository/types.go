package repository

import "time"

type User struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	DisplayName  string                 `json:"display_name"`
	PasswordHash string                 `json:"-"`
	Attrs        map[string]any         `json:"attrs"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type Pet struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Name      string         `json:"name"`
	Species   string         `json:"species"`
	Attrs     map[string]any `json:"attrs"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Habit struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	Icons     string    `json:"icons"`
	Cadence   string    `json:"cadence"` // "daily" | "everyN-<n_days>" | "weekly-<day_of_the_week>" or "weekly-<day1,day2,...>"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Game struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Title     string         `json:"title"`
	Status    string         `json:"status"`
	Attrs     map[string]any `json:"attrs"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
