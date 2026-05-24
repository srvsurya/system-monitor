package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(isDev bool) (*zap.Logger, error) {
	if isDev {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return cfg.Build()
	}
	return zap.NewProduction()
}
