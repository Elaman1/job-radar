package logger

import (
	"job-radar/internal/config"
	"log/slog"
	"os"
)

const (
	productionEnv = "prod"
)

func NewLogger(cfg config.Logger) *slog.Logger {
	var handler slog.Handler

	if cfg.Env == productionEnv {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	return slog.New(handler)
}
