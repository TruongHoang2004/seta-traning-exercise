// middleware/auth.go
package middleware

import (
	"collab-service/config"
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/external/user_service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware cho Gin framework
func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized: No Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		user, err := user_service.NewGraphQLClient(config.GetConfig().UserServiceEndpoint).ValidateToken(c.Request.Context(), tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Giữ đúng kiểu dữ liệu
		c.Set("userId", user.ID)
		c.Set("userRole", user.Role)

		c.Next()
	}
}

// GetUserInfoFromGin lấy userId và userRole từ Gin context
func GetUserInfoFromGin(c *gin.Context) (uuid.UUID, entity.UserType) {
	userId, _ := c.Get("userId")
	userRole, _ := c.Get("userRole")
	return userId.(uuid.UUID), userRole.(entity.UserType)
}
