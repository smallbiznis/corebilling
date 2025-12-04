package log

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides a zap logger configured for production.
var Module = fx.Provide(NewLogger)

// NewLogger returns a production zap logger and replaces globals.
func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(logger)
	return logger, nil
}
