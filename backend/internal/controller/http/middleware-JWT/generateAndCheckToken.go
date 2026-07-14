package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL = 72 * time.Hour

func GenerateToken(id int, login, role, jwtSecret string) (string, error) {
	now := time.Now()

	claims := Claims{
		Id: id,
		Login: login,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
