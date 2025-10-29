package app

import "os"

type Config struct {
	Port          string
	AllowedOrigin string
	JWTSecret     string
	DatabaseURL   string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" { origin = "*" }
	secret := os.Getenv("JWT_SECRET")
	if secret == "" { secret = "dev-secret-change-me" }
	dbURL := os.Getenv("DATABASE_URL")

	return Config{
		Port:          port,
		AllowedOrigin: origin,
		JWTSecret:     secret,
		DatabaseURL:   dbURL,
	}
}
