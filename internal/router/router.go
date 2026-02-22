package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/TradeLayers/BE/internal/handler"
	"github.com/TradeLayers/BE/internal/middleware"
)

func Setup(db *gorm.DB, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.ZapLogger(logger))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	healthHandler := handler.NewHealthHandler(db)

	api := r.Group("/api/v1")
	{
		api.GET("/health", healthHandler.Health)
	}

	return r
}
