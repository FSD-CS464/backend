package app

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"fsd-backend/internal/auth"
	"fsd-backend/internal/db"
	"fsd-backend/internal/middleware"
	"fsd-backend/internal/routers"
)

func NewServer(cfg Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg.AllowedOrigin))
	r.Use(middleware.Prometheus())

	signer := auth.NewSigner(cfg.JWTSecret,
		15*time.Minute,
		7*24*time.Hour,
	)

	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil { panic(err) }

	routers.RegisterSystemRoutes(r)
	routers.RegisterAPIV1(r, cfg, signer, pool)
	routers.RegisterWS(r, cfg, signer)

	return r
}
