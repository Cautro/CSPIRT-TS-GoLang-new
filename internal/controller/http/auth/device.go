package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	usersUsecase "cspirt/internal/usecase/user"
)

type DeviceTokenInput struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required"` // "ios" or "android"
}

func RegisterDeviceHandler(u *usersUsecase.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetString("Id") 
		if userIDStr == "" { c.JSON(401, gin.H{"error": "unauthorized"}); return }
		userID, err := strconv.Atoi(userIDStr)
		if err == nil { c.JSON(500, gin.H{"error": "Server error"}); return }

		var input DeviceTokenInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = u.RegisterDevice(c.Request.Context(), int64(userID), input.Token, input.Platform)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register device"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func UnregisterDeviceHandler(u *usersUsecase.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetString("Id") 
		if userIDStr == "" { c.JSON(401, gin.H{"error": "unauthorized"}); return }
		userID, err := strconv.Atoi(userIDStr)
		if err == nil { c.JSON(500, gin.H{"error": "Server error"}); return }

		var input struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = u.UnregisterDevice(c.Request.Context(), int64(userID), input.Token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unregister device"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}