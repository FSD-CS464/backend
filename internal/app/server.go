package app

import (
	"github.com/gin-gonic/gin"
	"fsd-backend/internal/middleware"
	"fsd-backend/internal/routers"
)

func NewServer(cfg Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg.AllowedOrigin))
	r.Use(middleware.Prometheus())

	routers.RegisterSystemRoutes(r)    // /healthz, /readyz, /metrics
	routers.RegisterAPIV1(r, cfg)      // /api/v1/*
	routers.RegisterWS(r, cfg)         // /ws

	return r
}
