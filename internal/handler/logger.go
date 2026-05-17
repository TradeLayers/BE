package handler

import (
	"github.com/TradeLayers/BE/internal/requestlog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func requestLogger(c *gin.Context) *zap.Logger {
	return requestlog.FromContext(c.Request.Context())
}
