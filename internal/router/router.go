package router

import (
	"os"

	"firebase.google.com/go/auth"
	appErrors "github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	appLogger "github.com/TradeLayers/BE/internal/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/TradeLayers/BE/internal/handler"
	"github.com/TradeLayers/BE/internal/middleware"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/service"
)

func Setup(db *gorm.DB, log *zap.Logger, authClient *auth.Client, finnhubClient finnhub.Client, priceMap *finnhub.PriceMap, wsClient *finnhub.WSClient) *gin.Engine {
	r := gin.New()
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		appErrors.ReturnError(c, appErrors.ErrInternal)
	}))
	r.Use(appLogger.HTTPMiddleware(log))

	frontendUrl := os.Getenv("FRONTEND_URL")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendUrl},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	healthHanlder := handler.NewHealthHandler(db)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	stockRepo := repository.NewStockRepository(db)
	stockService := service.NewStockService(finnhubClient, priceMap, stockRepo, wsClient)
	stockHandler := handler.NewStockHandler(stockService)

	holdingsRepo := repository.NewHoldingsRepository()
	transactionsRepo := repository.NewTransactionsRepository()
	portfolioService := service.NewPortfolioService(db, stockRepo, holdingsRepo, transactionsRepo, finnhubClient, priceMap, wsClient)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)

	watchlistRepo := repository.NewWatchlistRepository()
	watchlistService := service.NewWatchlistService(db, stockRepo, watchlistRepo, finnhubClient, priceMap, wsClient)
	watchlistHandler := handler.NewWatchlistHandler(watchlistService)
	notificationRepo := repository.NewNotificationRepository()
	notificationService := service.NewNotificationService(db, watchlistRepo, stockRepo, notificationRepo, finnhubClient, priceMap, wsClient)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	alertRepo := repository.NewAlertRepository()
	alertService := service.NewAlertService(db, stockRepo, alertRepo, finnhubClient, priceMap, wsClient)
	alertHandler := handler.NewAlertHandler(alertService)

	api := r.Group("/api")
	{
		api.GET("/health", healthHanlder.Health)

		protected := api.Group("")

		protected.Use(middleware.FirebaseAuth(authClient))

		protected.POST("/user", userHandler.CreateOrFetchUser)
		protected.PATCH("/user", userHandler.UpdateFields)
		protected.DELETE("/user", userHandler.DeleteUserAccount)

		protected.GET("/stocks", stockHandler.GetAllStocks)
		protected.GET("/stocks/quote/:symbol", stockHandler.GetQuote)
		protected.POST("/stocks/quotes", stockHandler.GetQuotes)
		protected.GET("/stocks/search", stockHandler.SearchStocks)
		protected.GET("/stocks/profile/:symbol", stockHandler.GetProfile)
		protected.GET("/stocks/candles", stockHandler.GetCandles)

		protected.GET("/portfolio/holdings", portfolioHandler.GetHoldings)
		protected.GET("/portfolio/transactions", portfolioHandler.GetTransactions)
		protected.GET("/portfolio/transactions.csv", portfolioHandler.ExportTransactionsCSV)
		protected.GET("/portfolio/history", portfolioHandler.GetHistory)
		protected.POST("/portfolio/buy", portfolioHandler.Buy)
		protected.POST("/portfolio/sell", portfolioHandler.Sell)

		protected.GET("/watchlist", watchlistHandler.List)
		protected.POST("/watchlist", watchlistHandler.Add)
		protected.DELETE("/watchlist/:symbol", watchlistHandler.Remove)
		protected.PATCH("/watchlist/:symbol/threshold", watchlistHandler.UpdateThreshold)

		protected.GET("/notifications/unread", notificationHandler.ListUnread)
		protected.PATCH("/notifications/:id/read", notificationHandler.MarkRead)

		protected.GET("/alerts", alertHandler.List)
		protected.POST("/alerts", alertHandler.Create)
		protected.DELETE("/alerts/:id", alertHandler.Delete)
	}

	return r
}
