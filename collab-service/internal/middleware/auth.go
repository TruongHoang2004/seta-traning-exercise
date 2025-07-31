// middleware/auth.go
package middleware

import (
	"collab-service/pkg/config"
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(config.GetConfig().JWTAccessSecret)

type contextKey string

const userCtxKey = contextKey("user_id")

// AuthMiddleware cho Gin framework
func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		ValidateToken(c, authHeader)
	}
}

// ValidateToken kiểm tra tính hợp lệ của token
func ValidateToken(c *gin.Context, authHeader string) {
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		c.Abort()
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		// Không có "Bearer " prefix
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
		c.Abort()
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is not valid"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		c.Abort()
		return
	}

	userID, ok := claims["userId"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid userId in token"})
		c.Abort()
		return
	}

	// Lưu userId vào context của Gin
	c.Set("userId", userID)

	// Cũng có thể lưu vào request context nếu cần
	ctx := context.WithValue(c.Request.Context(), userCtxKey, userID)
	c.Request = c.Request.WithContext(ctx)
	c.Next()

}

func extractOperationName(query string) string {
	query = strings.TrimSpace(query)
	lines := strings.Split(query, "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := strings.TrimSpace(lines[0])
	parts := strings.Fields(firstLine)

	// Ví dụ: mutation login { ... } hoặc query me { ... }
	if len(parts) >= 2 {
		return parts[1] // login hoặc renewToken
	}
	return ""
}

// GetUserIDFromGin lấy userId từ Gin context
func GetUserIDFromGin(c *gin.Context) (string, error) {
	userID, exists := c.Get("userId")
	if !exists {
		return "", errors.New("no user in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.New("invalid userId type")
	}

	return userIDStr, nil
}

// OptionalAuthMiddleware - middleware tùy chọn, không bắt buộc token
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Không có token, tiếp tục mà không set userId
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			// Token format không đúng, nhưng vẫn tiếp tục
			c.Next()
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secretKey), nil
		})

		// Nếu token hợp lệ, lưu user_id
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userID, ok := claims["user_id"].(string); ok {
					c.Set("user_id", userID)
				}
			}
		}

		c.Next()
	}
}

// ForContext lấy user_id từ standard context (nếu cần)
func ForContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userCtxKey).(string)
	if !ok {
		return "", errors.New("no user in context")
	}
	return userID, nil
}
