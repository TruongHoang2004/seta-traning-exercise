package jwt

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"user-service/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	accessSecret  = []byte(config.GetConfig().JWTAccessSecret)  // Nên lưu trong biến môi trường
	refreshSecret = []byte(config.GetConfig().JWTRefreshSecret) // Nên lưu trong biến môi trường
)

func getExpirationDuration(expiration string) (time.Duration, error) {
	if strings.HasSuffix(expiration, "d") {
		daysStr := strings.TrimSuffix(expiration, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, errors.New("invalid day duration format")
		}
		return time.Hour * 24 * time.Duration(days), nil
	}
	// fallback for standard durations like "15m", "2h45m", etc.
	return time.ParseDuration(expiration)
}

type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func CreateAccessToken(userID string) (string, error) {
	accessExpiration, err := getExpirationDuration(config.GetConfig().JWTAccessExpiration)
	if err != nil {
		return "", err
	}
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessSecret)
}

func CreateRefreshToken(userID string) (string, error) {
	refreshExpiration, err := getExpirationDuration(config.GetConfig().JWTRefreshExpiration)
	if err != nil {
		return "", err
	}
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecret)
}

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
