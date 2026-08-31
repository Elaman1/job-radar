package handlers

import (
	"github.com/gin-gonic/gin"
	"job-radar/internal/databases/postgres"
	"net/http"
)

type HealthHandler struct {
	db postgres.Postgres
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthLive(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "job-radar",
	})
}
