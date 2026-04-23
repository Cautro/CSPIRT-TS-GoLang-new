package users

import (
	"cspirt/internal/repo"
	"log/slog"
	"net/http"
	"github.com/gin-gonic/gin"
)

type UsersService struct {
	users     repo.UserRepository
	log 	  *slog.Logger
}

func NewUsersService(users repo.UserRepository, jwtSecret string) *UsersService {
	return &UsersService{
		users:     users,
		log:       slog.Default(),
	}
}

func (s *UsersService) GetUsersHandlerService() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := s.users.GetAllUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}
		c.JSON(http.StatusOK, users)
}}