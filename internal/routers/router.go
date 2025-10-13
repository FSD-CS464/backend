package routers

import (
	"github.com/gin-gonic/gin"
	"fsd-backend/internal/controllers"
	"fsd-backend/internal/middleware"
)

type cfgLike interface {
	// we only need AllowedOrigin or JWTSecret here later if desired
}

func RegisterSystemRoutes(r *gin.Engine) {
	r.GET("/healthz", controllers.HealthCheck)
	r.GET("/readyz", controllers.HealthCheck) // later: ping DB/Redis
	r.GET("/metrics", middleware.MetricsHandler())
}

func RegisterAPIV1(r *gin.Engine, cfg cfgLike) {
	v1 := r.Group("/api/v1")

	// public
	v1.POST("/auth/login", controllers.LoginStub)
	v1.POST("/auth/refresh", controllers.RefreshStub)

	// protected
	auth := v1.Group("/")
	auth.Use(middleware.JWT())
	{
		auth.GET("/users", controllers.GetUsers)
		auth.GET("/users/:id", controllers.GetUserByID)
		auth.POST("/users", controllers.CreateUser)
		auth.PUT("/users/:id", controllers.UpdateUser)
		auth.DELETE("/users/:id", controllers.DeleteUser)
	}
}

func RegisterWS(r *gin.Engine, cfg cfgLike) {
	r.GET("/ws", controllers.WSHandler) // token can be query param for now
}
