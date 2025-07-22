// middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = getSecretKey()

func getSecretKey() string {
	key := os.Getenv("JWT_SECRET_KEY")
	if key == "" {
		// Fallback to a default key only in development
		key = "your_secret_key"
		log.Println("Warning: Using default JWT secret key. Set JWT_SECRET_KEY environment variable in production.")
	}
	return key
}

type contextKey string

const userCtxKey = contextKey("user_id")

// AuthMiddleware cho Gin framework
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
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

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
			c.Abort()
			return
		}

		// Lưu user_id vào context của Gin
		c.Set("user_id", userID)

		// Cũng có thể lưu vào request context nếu cần
		ctx := context.WithValue(c.Request.Context(), userCtxKey, userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetUserIDFromGin lấy user_id từ Gin context
func GetUserIDFromGin(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("no user in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.New("invalid user_id type")
	}

	return userIDStr, nil
}

// OptionalAuthMiddleware - middleware tùy chọn, không bắt buộc token
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Không có token, tiếp tục mà không set user_id
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
