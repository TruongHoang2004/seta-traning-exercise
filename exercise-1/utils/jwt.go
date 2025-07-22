package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key")

func GenerateJWT(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
