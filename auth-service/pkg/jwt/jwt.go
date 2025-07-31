package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	accessSecret  = []byte("your_access_secret_key")  // Nên lưu trong biến môi trường
	refreshSecret = []byte("your_refresh_secret_key") // Nên lưu trong biến môi trường
)

type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// CreateAccessToken generates a JWT access token valid for 15 minutes
func CreateAccessToken(userID string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessSecret)
}

// CreateRefreshToken generates a JWT refresh token valid for 7 days
func CreateRefreshToken(userID string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecret)
}

// ValidateToken verifies a token and returns the user ID if valid
func ValidateToken(tokenStr string, isAccessToken bool) (string, error) {
	secret := accessSecret
	if !isAccessToken {
		secret = refreshSecret
	}

	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}

// RefreshTokens validates a refresh token and returns new access and refresh tokens
func RefreshTokens(refreshToken string) (string, string, error) {
	// Validate the refresh token
	userID, err := ValidateToken(refreshToken, false)
	if err != nil {
		return "", "", err
	}

	// Generate new access token
	accessToken, err := CreateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	// Generate new refresh token
	newRefreshToken, err := CreateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
