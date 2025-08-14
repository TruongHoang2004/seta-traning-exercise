// middleware/auth.go
package middleware

import (
	"collab-service/config"
	"collab-service/pkg/client"
	"context"
	"errors"
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
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized: No Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		ValidateToken(c, tokenStr)
	}
}

// ValidateToken kiểm tra tính hợp lệ của token
func ValidateToken(c *gin.Context, authHeader string) {

	user, err := client.NewGraphQLClient(config.GetConfig().UserServiceEndpoint).ValidateToken(c.Request.Context(), authHeader)
	if err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Giữ đúng kiểu dữ liệu
	c.Set("userId", user.UserID)
	c.Set("userRole", user.Role)

	c.Next()
}

// GetUserInfoFromGin lấy userId và userRole từ Gin context
func GetUserInfoFromGin(c *gin.Context) (string, client.UserType) {
	userID, exists := c.Get("userId")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized: No user in context"})
		return "", ""
	}

	userRole, exists := c.Get("userRole")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized: No user role in context"})
		return "", ""
	}

	// Ép kiểu về UserType (hoặc UserRole nếu bạn dùng tên đó)
	userIDStr, ok1 := userID.(string)
	userRoleTyped, ok2 := userRole.(client.UserType)

	if !ok1 || !ok2 {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized: Invalid user info type"})
		return "", ""
	}

	return userIDStr, userRoleTyped
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
