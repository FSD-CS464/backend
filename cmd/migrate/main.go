package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func main() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is empty")
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "database/migrations"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

    if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

    // Basic arg handling to support reset/down/redo in addition to up
    // Usage examples:
    //   go run ./cmd/migrate            # up
    //   go run ./cmd/migrate -reset     # drops all migrations then re-applies up
    //   go run ./cmd/migrate -down      # steps down one migration
    //   go run ./cmd/migrate -down-all  # steps down all migrations
    //   go run ./cmd/migrate -redo      # redo last migration
    args := os.Args[1:]
    var op string
    if len(args) > 0 {
        op = args[0]
    }

    switch op {
    case "-reset":
        if err := goose.Reset(db, migrationsDir); err != nil {
            log.Fatal(err)
        }
        if err := goose.Up(db, migrationsDir); err != nil {
            log.Fatal(err)
        }
        log.Println("migrations reset and re-applied")
    case "-down":
        if err := goose.Down(db, migrationsDir); err != nil {
            log.Fatal(err)
        }
        log.Println("stepped down one migration")
    case "-down-all":
        if err := goose.DownTo(db, migrationsDir, 0); err != nil {
            log.Fatal(err)
        }
        log.Println("stepped down all migrations")
    case "-redo":
        if err := goose.Redo(db, migrationsDir); err != nil {
            log.Fatal(err)
        }
        log.Println("redo last migration applied")
    default:
        if err := goose.Up(db, migrationsDir); err != nil {
            log.Fatal(err)
        }
        log.Println("migrations applied")
    }
}
