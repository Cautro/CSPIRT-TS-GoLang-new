package handlers

import (
	"cspirt/internal/models"
	sr "cspirt/internal/service/notes"
	"cspirt/internal/storage"
	"cspirt/internal/logger"

	"github.com/gin-gonic/gin"
)

func GetNotesHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		notes := sr.NewNoteService(s, s.Secret)
		result, err := notes.GetAllNotes()
		if err != nil || result == nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}

		c.JSON(200, gin.H{"All notes": result})
	}
}

func AddNoteHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		notes := sr.NewNoteService(s, s.Secret)
		user, err := s.GetUserByLogin(login)
		if err != nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "notes",
				Message: "server error: " + err.Error(),
			})
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}

		var in *models.AddNewNoteResponse
		if err := c.ShouldBindJSON(&in); err != nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "notes",
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		needUser := &models.SafeUser{
			ID: user.ID,
			Name: user.Name,
			LastName: user.LastName,
			FullName: user.FullName,
			Login: user.Login,
			Rating: user.Rating,
			Role: user.Role,
			Class: user.Class,
		} 

		result := notes.AddNewNote(login, in, needUser)
		if err != nil || result == nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}

		writeLog(logger.LogEntry{
			Level: "info",
			Login: login,
			Class: user.Class,
			Role: user.Role,
			Action: "added_note",
			Message: "Added new note",
		})
		c.JSON(200, gin.H{"Add notes": result})
	}
}

func DeleteNoteHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for deleting a note
	}
}