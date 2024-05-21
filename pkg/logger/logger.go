package slogger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func GetLogger() *slog.Logger {

	var lvl = new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
	return logger
}
