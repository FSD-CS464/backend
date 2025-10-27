package app

import (
	"time"

	"github.com/gin-gonic/gin"

	"fsd-backend/internal/auth"
	"fsd-backend/internal/middleware"
	"fsd-backend/internal/routers"
)

func NewServer(cfg Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg.AllowedOrigin))
	r.Use(middleware.Prometheus())

	// Minimal sane TTLs (env-configurable later)
	signer := auth.NewSigner(cfg.JWTSecret,
		15*time.Minute, // access
		7*24*time.Hour, // refresh
	)

	routers.RegisterSystemRoutes(r)
	routers.RegisterAPIV1(r, cfg, signer)
	routers.RegisterWS(r, cfg)

	return r
}
