package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// only log errors and slow requests
		status := c.Writer.Status()
		took := time.Since(start)

		if status >= 500 {
			log.Error("request failed",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", status),
				zap.Duration("took", took),
			)
		} else if took > 500*time.Millisecond {
			log.Warn("slow request",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Duration("took", took),
			)
		}
		// 2xx, 4xx — no log. handler already logged if needed.
	}
}
