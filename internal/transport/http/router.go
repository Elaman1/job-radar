package http

import (
	"github.com/gin-gonic/gin"
	"job-radar/internal/transport/http/handlers"
	"job-radar/internal/transport/http/middleware"
	"log/slog"
)

func NewRouter(
	log *slog.Logger,
	healthHandler *handlers.HealthHandler,
	checkerHandler *handlers.CheckerHandler,
) *gin.Engine {
	router := gin.New()

	router.Use(
		middleware.Logger(log),
		gin.Recovery(),
	)

	api := router.Group("/api")

	// routes
	api.GET("/health/live", healthHandler.HealthLive)
	api.GET("/checker", checkerHandler.CheckerHandler)

	return router
}
