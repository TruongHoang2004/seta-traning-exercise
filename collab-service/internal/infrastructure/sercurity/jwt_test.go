package sercurity

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJWTAccessToken(t *testing.T) {
	// Test case 1: Valid token generation
	userID := "user123"
	role := "admin"
	token, err := GenerateJWTAccessToken(userID, role)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed back
	parsedToken, err := ParseJWTAccessToken(token)
	assert.NoError(t, err)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["userId"])
	assert.Equal(t, role, claims["role"])

	// Check expiration time (should be ~30 minutes in the future)
	exp, ok := claims["exp"].(float64)
	assert.True(t, ok)
	expTime := time.Unix(int64(exp), 0)
	assert.True(t, expTime.After(time.Now()))
	assert.True(t, expTime.Before(time.Now().Add(31*time.Minute)))
}

func TestGenerateJWTRefreshToken(t *testing.T) {
	// Test valid token generation
	userID := "user456"
	token, err := GenerateJWTRefreshToken(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token format (should be 3 parts separated by dots)
	parts := strings.Split(token, ".")
	assert.Equal(t, 3, len(parts))

	// Verify token can be parsed back
	parsedUserID, err := ParseJWTRefreshToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)
}

func TestRenewToken(t *testing.T) {
	// Generate initial refresh token
	userID := "user789"
	refreshToken, err := GenerateJWTRefreshToken(userID)
	assert.NoError(t, err)

	// Renew token
	newAccessToken, newRefreshToken, err := RenewToken(refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)

	// Verify new access token
	parsedAccessToken, err := ParseJWTAccessToken(newAccessToken)
	assert.NoError(t, err)
	accessClaims, ok := parsedAccessToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, accessClaims["userId"])

	// Verify new refresh token
	parsedUserID, err := ParseJWTRefreshToken(newRefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)
}
