package handlers

import (
	"github.com/gin-gonic/gin"
	// "cspirt/internal/storage"
	// "cspirt/internal/logger"
)

// HealthHandler returns the service health status.
// @Summary Health check
// @Description Returns a simple health response.
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
