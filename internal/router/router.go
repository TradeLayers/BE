package router

import (
	"net/url"
	"os"
	"strings"

	"firebase.google.com/go/auth"
	appErrors "github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/config"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/requestlog"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/TradeLayers/BE/internal/handler"
	"github.com/TradeLayers/BE/internal/middleware"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/service"
)

var localFrontendOrigins []string = []string{
	"http://localhost:3000",
	"http://localhost:3001",
	"http://localhost:3002",
	"http://127.0.0.1:3000",
	"http://127.0.0.1:3001",
	"http://127.0.0.1:3002",
}

func Setup(db *gorm.DB, log *zap.Logger, authClient *auth.Client, finnhubClient finnhub.Client, priceMap *finnhub.PriceMap, wsClient *finnhub.WSClient, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestlog.FromContext(c.Request.Context()).Error("panic recovered", zap.Any("recovered", recovered))
		appErrors.ReturnError(c, appErrors.ErrInternal)
	}))
	r.Use(requestlog.HTTPMiddleware(log))

	allowedOrigins := buildAllowedOrigins(os.Getenv("FRONTEND_URL"))

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return isAllowedOrigin(origin, allowedOrigins)
		},
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

	donationService := service.NewDonationService(cfg.StripeSecretKey, cfg.StripeDonationCurrency, cfg.StripeDonationProductName)
	donationHandler := handler.NewDonationHandler(donationService, allowedOrigins)

	api := r.Group("/api")
	{
		api.GET("/health", healthHanlder.Health)
		api.POST("/donations/checkout", donationHandler.CreateCheckoutSession)

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

func buildAllowedOrigins(frontendURLs string) []string {
	seen := make(map[string]struct{})
	allowedOrigins := make([]string, 0)

	addOrigin := func(origin string) {
		trimmedOrigin := strings.TrimSpace(origin)
		if trimmedOrigin == "" {
			return
		}

		if _, exists := seen[trimmedOrigin]; exists {
			return
		}

		seen[trimmedOrigin] = struct{}{}
		allowedOrigins = append(allowedOrigins, trimmedOrigin)
	}

	for _, origin := range strings.Split(frontendURLs, ",") {
		addOrigin(origin)
	}

	if frontendURLs == "" || containsLocalOrigin(allowedOrigins) {
		for _, origin := range localFrontendOrigins {
			addOrigin(origin)
		}
	}

	return allowedOrigins
}

func containsLocalOrigin(origins []string) bool {
	for _, origin := range origins {
		parsedOrigin, err := url.Parse(origin)
		if err != nil {
			continue
		}

		switch parsedOrigin.Hostname() {
		case "localhost", "127.0.0.1":
			return true
		}
	}

	return false
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}
