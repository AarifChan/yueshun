package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger := log.Info()
		if status >= 400 {
			logger = log.Warn()
		}
		if status >= 500 {
			logger = log.Error()
		}

		logger.
			Str("client_ip", clientIP).
			Str("method", method).
			Str("path", path).
			Str("query", query).
			Int("status", status).
			Dur("latency", latency).
			Str("error", c.Errors.ByType(gin.ErrorTypePrivate).String()).
			Msg("HTTP request")
	}
}
