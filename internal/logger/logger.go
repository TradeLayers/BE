package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TradeLayers/BE/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg *config.Config) (*zap.Logger, error) {
	if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	infoFile, err := os.OpenFile(
		filepath.Join(cfg.LogDir, cfg.InfoLogFile),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return nil, fmt.Errorf("open info log file: %w", err)
	}

	errorFile, err := os.OpenFile(
		filepath.Join(cfg.LogDir, cfg.ErrorLogFile),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		_ = infoFile.Close()
		return nil, fmt.Errorf("open error log file: %w", err)
	}

	level := parseLevel(cfg.LogLevel)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	infoCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(infoFile),
		zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= level && l < zapcore.ErrorLevel
		}),
	)
	errorCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(errorFile),
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
