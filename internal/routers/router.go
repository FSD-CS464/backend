package routers

import (
	"github.com/gin-gonic/gin"

	"fsd-backend/internal/auth"
	"fsd-backend/internal/controllers"
	"fsd-backend/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

type cfgLike interface{}

func RegisterSystemRoutes(r *gin.Engine) {
	r.GET("/health", controllers.HealthCheck)
	r.GET("/ready", controllers.HealthCheck)
	r.GET("/metrics", middleware.MetricsHandler())
}

func RegisterAPIV1(r *gin.Engine, cfg cfgLike, signer *auth.Signer, pool *pgxpool.Pool) {
	v1 := r.Group("/api/v1")

	authCtl := controllers.NewAuthController(signer, pool)
	jwtmw := middleware.NewJWT(signer)

	// public
	v1.POST("/auth/register", authCtl.Register)
	v1.POST("/auth/login", authCtl.Login)
	v1.POST("/auth/refresh", authCtl.Refresh)

	// protected
	protected := v1.Group("/")
	protected.Use(jwtmw.Require())
	{
		protected.GET("/auth/me", authCtl.Me)

		udb := controllers.NewUserController(pool)
		protected.GET("/users", udb.List)
		protected.GET("/users/:id", udb.GetByID)
		protected.POST("/users", udb.Create)
		protected.PUT("/users/:id/name", udb.UpdateName)
		protected.DELETE("/users/:id", udb.Delete)

		// pdb := controllers.NewPetControllerDB(pool)
		// protected.GET("/pets/:id", pdb.GetByID)
		// protected.POST("/pets", pdb.Create)

		hdb := controllers.NewHabitController(pool)
		protected.GET("/habits", hdb.List)
		protected.GET("/habits/:id", hdb.GetByID)
		protected.POST("/habits", hdb.Create)
		protected.PUT("/habits/:id", hdb.Update)
		protected.DELETE("/habits/:id", hdb.Delete)

		// gdb := controllers.NewGameControllerDB(pool)
		// protected.GET("/games/:id", gdb.GetByID)
		// protected.POST("/games", gdb.Create)

		// Game API endpoints for Godot game
		gameCtl := controllers.NewGameController(pool)
		gameGroup := protected.Group("/game")
		{
			gameGroup.GET("/data", gameCtl.GetUserData)
			gameGroup.POST("/save", gameCtl.SaveGameData)
		}
	}
}

func RegisterWS(r *gin.Engine, cfg cfgLike) {
	r.GET("/ws", controllers.WSHandler)
}
