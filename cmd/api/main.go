package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/TradeLayers/BE/internal/config"
	"github.com/TradeLayers/BE/internal/database"
	"github.com/TradeLayers/BE/internal/router"
	"github.com/TradeLayers/BE/internal/server"
)

func main() {
	cfg := config.Load()

	logger, err := zap.NewDevelopment()
	if cfg.LogLevel == "production" {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	db, err := database.Connect(cfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	logger.Info("database connected")

	r := router.Setup(db, logger)

	server.Run(cfg.AppPort, r, logger)
}
