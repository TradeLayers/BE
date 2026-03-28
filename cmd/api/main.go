package main

import (
	"context"
	"log"

	"github.com/TradeLayers/BE/internal/config"
	"github.com/TradeLayers/BE/internal/database"
	"github.com/TradeLayers/BE/internal/logger"
	"github.com/TradeLayers/BE/internal/router"
	"github.com/TradeLayers/BE/internal/server"
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

	r := router.Setup(db, zapLogger, authClient)

	server.Run(cfg.AppPort, r, zapLogger)
}
