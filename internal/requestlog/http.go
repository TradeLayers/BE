package requestlog

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func HTTPMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		requestLog := log.With(zap.String(RequestIDKey, requestID))
		requestCtx := WithRequestID(WithLogger(c.Request.Context(), requestLog), requestID)
		c.Request = c.Request.WithContext(requestCtx)
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		if query != "" {
			path = path + "?" + query
		}

		requestLog.Info("http_request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
			zap.Int("bytes_in", int(c.Request.ContentLength)),
			zap.Int("bytes_out", c.Writer.Size()),
		)
	}
}
