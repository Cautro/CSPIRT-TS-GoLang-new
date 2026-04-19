package handlers

import (
	"github.com/gin-gonic/gin"
	// "cspirt/internal/storage"
	// "cspirt/internal/logger"
)

func HealthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}