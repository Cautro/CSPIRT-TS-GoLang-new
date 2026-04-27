package handlers

import (
	"cspirt/internal/storage"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetNotesHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for getting notes
	}
}

func AddNoteHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for adding a note
	}
}

func DeleteNoteHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for deleting a note
	}
}