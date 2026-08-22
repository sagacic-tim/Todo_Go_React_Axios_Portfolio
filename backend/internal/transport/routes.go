// backend/internal/transport/routes.go
package transport

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	// Initialise shared rate-limiter instances.
	initRateLimiters()

	r := gin.Default()
	r.SetTrustedProxies(nil)

	// ── Global middleware (every route) ────────────────────────────────────
	r.Use(SecurityHeaders())
	r.Use(globalLimiter.Middleware())

	api := r.Group("/api")

	// ── Public health check (no auth, no rate limit — used by deploy CI) ───
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ── Public auth routes (stricter rate limit, no session required) ──────
	auth := api.Group("/auth")
	auth.Use(authLimiter.Middleware())
	{
		auth.POST("/signup", makeSignup(db))
		auth.POST("/login",  makeLogin(db))
		auth.POST("/logout", makeLogout())
	}

	// ── Protected auth routes ──────────────────────────────────────────────
	authProtected := api.Group("/auth")
	authProtected.Use(authLimiter.Middleware())
	authProtected.Use(requireAuth(db))
	{
		authProtected.GET("/me", makeMe())
	}

	// ── Protected task routes ──────────────────────────────────────────────
	tasks := api.Group("/tasks")
	tasks.Use(requireAuth(db))
	{
		tasks.GET   ("",     makeGetTasks(db))
		tasks.POST  ("",     makeCreateTask(db))
		tasks.PATCH ("/:id", makeUpdateTask(db))
		tasks.DELETE("/:id", makeDeleteTask(db))
	}

	// ── Static frontend ────────────────────────────────────────────────────
	r.StaticFile("/", "./frontend/dist/index.html")
	r.Static("/static", "./frontend/dist/assets")

	return r
}
