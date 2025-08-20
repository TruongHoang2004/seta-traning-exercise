package middleware

import (
	"collab-service/internal/infrastructure/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerpMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)

		logger.Info("Request handled",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", latency,
		)

	}
}
