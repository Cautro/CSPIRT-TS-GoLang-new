package utils

import "github.com/golang-jwt/jwt/v5"

func GenerateToken(login, jwtSecret string) (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	"Login": login,
	}).SignedString([]byte(jwtSecret))

	if err != nil {
		return "", err
	}

	return token, nil
}