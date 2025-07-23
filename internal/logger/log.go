package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger = zap.NewNop()

// Initialize initiates global logger
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}
