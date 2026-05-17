package main

import (
	"context"
	"log"

	"github.com/TradeLayers/BE/internal/config"
	"github.com/TradeLayers/BE/internal/database"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/logger"
	"github.com/TradeLayers/BE/internal/router"
	"github.com/TradeLayers/BE/internal/server"
	"github.com/TradeLayers/BE/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	zapLogger, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer zapLogger.Sync() //nolint:errcheck // best-effort flush

	app, err := config.InitFirebase()
	if err != nil {
		zapLogger.Fatal("failed to initialize firebase app", zap.Error(err))
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		zapLogger.Fatal("failed to initialize firebase auth client", zap.Error(err))
	}

	db, err := database.Connect(cfg)
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	zapLogger.Info("database connected")

	priceMap := finnhub.NewPriceMap()
	finnhubClient := finnhub.NewClient(cfg.FinnhubAPIKey)
	wsClient := finnhub.NewWSClient(cfg.FinnhubAPIKey, cfg.FinnhubWSURL, priceMap, zapLogger)

	wsClient.Subscribe(service.DefaultSymbols())

	ctx, cancelWS := context.WithCancel(context.Background())
	defer cancelWS()
	go wsClient.Run(ctx)

	r := router.Setup(db, zapLogger, authClient, finnhubClient, priceMap, wsClient, cfg)

	server.Run(cfg.AppPort, r, zapLogger)
}
