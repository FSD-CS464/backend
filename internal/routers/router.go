package routers

import (
	"github.com/gin-gonic/gin"

	"fsd-backend/internal/auth"
	"fsd-backend/internal/controllers"
	"fsd-backend/internal/middleware"
)

type cfgLike interface{}

func RegisterSystemRoutes(r *gin.Engine) {
	r.GET("/health", controllers.HealthCheck)
	r.GET("/ready", controllers.HealthCheck)
	r.GET("/metrics", middleware.MetricsHandler())
}

// NOTE: accept signer so we can wire JWT + auth handlers
func RegisterAPIV1(r *gin.Engine, cfg cfgLike, signer *auth.Signer) {
	v1 := r.Group("/api/v1")

	authCtl := controllers.NewAuthController(signer)
	jwtmw := middleware.NewJWT(signer)

	// public
	v1.POST("/auth/login", authCtl.Login)
	v1.POST("/auth/refresh", authCtl.Refresh)

	// protected
	protected := v1.Group("/")
	protected.Use(jwtmw.Require())
	{
		protected.GET("/auth/me", authCtl.Me)

		protected.GET("/users", controllers.GetUsers)
		protected.GET("/users/:id", controllers.GetUserByID)
		protected.POST("/users", controllers.CreateUser)
		protected.PUT("/users/:id", controllers.UpdateUser)
		protected.DELETE("/users/:id", controllers.DeleteUser)
	}
}

func RegisterWS(r *gin.Engine, cfg cfgLike) {
	// keep as-is for now (you can validate token in query later if needed)
	r.GET("/ws", controllers.WSHandler)
}
