package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TradeLayers/BE/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(cfg *config.Config) (*zap.Logger, error) {
	if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	level := parseLevel(cfg.LogLevel)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	infoLogSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, cfg.InfoLogFile),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})

	errorLogSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, cfg.ErrorLogFile),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})

	infoCore := zapcore.NewCore(
		encoder,
		infoLogSyncer,
		zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= level && l < zapcore.ErrorLevel
		}),
	)

	errorCore := zapcore.NewCore(
		encoder,
		errorLogSyncer,
		zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= zapcore.ErrorLevel
		}),
	)

	return zap.New(zapcore.NewTee(infoCore, errorCore), zap.AddCaller()), nil
}

func parseLevel(logLevel string) zapcore.Level {
	switch strings.ToLower(logLevel) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
