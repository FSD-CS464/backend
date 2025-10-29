package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"fsd-backend/internal/app"
	mydb "fsd-backend/internal/db"
	"fsd-backend/internal/seed"
)

func main() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	cfg := app.LoadConfig()
	ctx := context.Background()

	pool, err := mydb.Connect(ctx, cfg.DatabaseURL)
	if err != nil { log.Fatal(err) }
	defer pool.Close()

	// Sample data — tweak freely
	data := seed.Data{
		Users: []seed.User{
			{Email: "alice@example.com", DisplayName: "Alice", Attrs: map[string]any{"role":"student","theme":"pink"}},
			{Email: "bob@example.com",   DisplayName: "Bob",   Attrs: map[string]any{"role":"student","theme":"dark"}},
			{Email: "charlie@example.com", DisplayName: "Charlie", Attrs: map[string]any{"role":"ta"}},
		},
		Pets: []seed.Pet{
			{UserEmail: "alice@example.com", Name: "Milo", Species: "cat", Attrs: map[string]any{"color":"white"}},
			{UserEmail: "bob@example.com",   Name: "Bolt", Species: "dog", Attrs: map[string]any{"age":2}},
		},
		Habits: []seed.Habit{
			{UserEmail: "alice@example.com", Title: "Daily Pomodoro", Cadence: "daily", Attrs: map[string]any{"duration":25}},
			{UserEmail: "bob@example.com",   Title: "Gym 3x",         Cadence: "3x/week"},
		},
		Games: []seed.Game{
			{UserEmail: "alice@example.com", Title: "Zelda",   Status: "playing", Attrs: map[string]any{"platform":"Switch"}},
			{UserEmail: "bob@example.com",   Title: "Hades 2", Status: "backlog", Attrs: map[string]any{"platform":"PC"}},
		},
	}

	if err := seed.Run(ctx, pool, data); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
	log.Println("seed complete")
}
