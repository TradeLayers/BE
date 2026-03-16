package router

import (
	"firebase.google.com/go/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/TradeLayers/BE/internal/handler"
	"github.com/TradeLayers/BE/internal/middleware"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/service"
)

func Setup(db *gorm.DB, logger *zap.Logger, authClient *auth.Client) *gin.Engine {
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
	portfolioRepo := repository.NewPortfolioRepository(db)
	portfolioService := service.NewPortfolioService(portfolioRepo)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)

	api := r.Group("/api")
	{
		api.GET("/health", healthHandler.Health)

		protected := api.Group("/")
		protected.Use(middleware.FirebaseAuth(authClient))

		protected.GET("/user", handler.GetPortfolio)
		protected.POST("/portfolios", portfolioHandler.CreatePortfolio)
	}

	return r
}
