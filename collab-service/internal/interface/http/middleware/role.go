// middleware/auth.go
package middleware

import (
	"collab-service/internal/domain/entity"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware cho Gin framework
func RoleMiddleware(role entity.UserType) gin.HandlerFunc {

	return func(c *gin.Context) {
		_, userRole := GetUserInfoFromGin(c)

		if userRole != role {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden: Insufficient permissions"})
			return
		}

		c.Next()
	}
}
