package utils

import (
	"fmt"
	"seta-training-exercise-1/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtAccessTokenSecret []byte
var jwtRefreshTokenSecret []byte

func init() {
	jwtAccessTokenSecret = []byte(config.GetConfig().JWTAccessSecret)
	jwtRefreshTokenSecret = []byte(config.GetConfig().JWTRefreshSecret)
}

func GenerateJWTAccessToken(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"exp":    time.Now().Add(time.Minute * 30).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtAccessTokenSecret)
}

func ParseJWTAccessToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return jwtAccessTokenSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

func GenerateJWTRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(time.Hour * 24 * 14).Unix(), // Refresh token valid for 14 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtRefreshTokenSecret)
}

func ParseJWTRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return jwtRefreshTokenSecret, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("couldn't parse claims")
	}

	// Return token only if we can find the userID
	userID, ok := claims["userId"].(string)
	if !ok {
		return "", fmt.Errorf("couldn't find userId in claims")
	}
	if userID == "" {
		return "", fmt.Errorf("userId is empty in claims")
	}
	return userID, nil
}

func RenewToken(token string) (string, string, error) {
	// Parse the refresh token to get userID
	userID, err := ParseJWTRefreshToken(token)
	if err != nil {
		return "", "", err
	}

	// Generate a new access token
	newAccessToken, err := GenerateJWTAccessToken(userID, "user") // Assuming role is "user" for simplicity
	if err != nil {
		return "", "", err
	}

	// Generate a new refresh token
	newRefreshToken, err := GenerateJWTRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
