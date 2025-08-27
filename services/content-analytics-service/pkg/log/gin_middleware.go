package log

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinZap(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		l.Info("http_request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ua", c.Request.UserAgent()),
		)
	}
}