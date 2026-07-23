package globalevent

import (
	entity "cspirt/internal/domain/globalEvent"
	usecase "cspirt/internal/usecase/globalEvent"
	permission "cspirt/internal/controller/permission/usecase"
	log "cspirt/pkg/logger"
	
	"time"
	"net/http"
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetGlobalEvents(usecase *usecase.GlobalEventUsecase) gin.HandlerFunc {
	return func(c *gin.Context)  {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		output, err := usecase.GetAllGlobalEvents(ctx)
		if err != nil {
			c.JSON(500, gin.H{"error": "Server error"})
			log.WriteSafe(log.LogEntry{Message: "Failed trying to get all global event"})
			return 
		}

		c.JSON(200, output)
	}
}

func AddInfoGlobalEvent(usecase *usecase.GlobalEventUsecase, perm permission.Usecase) gin.HandlerFunc {
	return func(c *gin.Context)  {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var input entity.GlobalEventInfoDTO
		if err := c.ShouldBindJSON(&input); err != nil { c.JSON(400, gin.H{"error":"Bad request"}); return }
		if err := usecase.AddInfoGlobalEvent(ctx, input, perm, c.GetString("Login")); err != nil { c.JSON(500, gin.H{"error":"Server error"}); return }

		c.JSON(200, gin.H{"status":"ok"})
	}
}

func AddQuizGlobalEvent(usecase *usecase.GlobalEventUsecase, perm permission.Usecase) gin.HandlerFunc {
	return func(c *gin.Context)  {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var input entity.GlobalEventQuizDTO
		if err := c.ShouldBindJSON(&input); err != nil { c.JSON(400, gin.H{"error":"Bad request"}); return }
		if err := usecase.AddQuizGlobalEvent(ctx, input, perm, c.GetString("Login")); err != nil { c.JSON(500, gin.H{"error":"Server error"}); return }

		c.JSON(200, gin.H{"status":"ok"})
	}
}

func DeleteInfoGlobalEvent(usecase *usecase.GlobalEventUsecase, perm permission.Usecase) gin.HandlerFunc {
	return func(c *gin.Context)  {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		idStr := c.Query("Id"); if idStr == "" { c.JSON(400, gin.H{"error":"Bad request"}); return }
		id, err := strconv.Atoi(idStr); if err != nil { c.JSON(500, gin.H{"error":"Server error"}); return }

		if err := usecase.DeleteInfoGlobalEvent(ctx, id, perm, c.GetString("Login")); err != nil { c.JSON(500, gin.H{"error":"Server error"}); return }
	
		c.JSON(200, gin.H{"status":"ok"})
	}
}

func DeleteQuizGlobalEvent(usecase *usecase.GlobalEventUsecase, perm permission.Usecase) gin.HandlerFunc {
	return func(c *gin.Context)  {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		idStr := c.Query("Id"); if idStr == "" { c.JSON(400, gin.H{"error":"Bad request"}); return }
		id, err := strconv.Atoi(idStr); if err != nil { c.JSON(500, gin.H{"error":"Server error"}); return }

		if err := usecase.DeleteQuizGlobalEvent(ctx, id, perm, c.GetString("Login")); err != nil { c.JSON(500, gin.H{"error":"Server error"}); return }
	
		c.JSON(200, gin.H{"status":"ok"})
	}
}

func Vote(usecase *usecase.GlobalEventUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		userID, ok := getUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user_id not found in context"})
			return
		}

		quizIdStr := c.Param("eventId")
		quizID, err := strconv.Atoi(quizIdStr)
		if err != nil || quizID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid eventId parameter"})
			return
		}

		var input entity.VoteToPutinDTO
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := usecase.Vote(ctx, userID, quizID, input.VoteItemId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func getUserID(c *gin.Context) (int, bool) {
	keys := []string{"Id", "userId", "user_id", "id"}
	for _, key := range keys {
		if val, exists := c.Get(key); exists {
			switch v := val.(type) {
			case int:
				return v, true
			case int64:
				return int(v), true
			case float64:
				return int(v), true
			case string:
				if id, err := strconv.Atoi(v); err == nil {
					return id, true
				}
			}
		}
	}
	return 0, false
}