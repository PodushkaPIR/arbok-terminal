package logging

import (
	"log/slog"
	"os"
	"strings"
)

var (
	globalLogger *slog.Logger
	logLevel    = new(slog.LevelVar)
)

func Init(level string) {
	var l slog.Level
	switch strings.ToLower(level) {
	case "trace", "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn", "warning":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelWarn
	}

	logLevel.Set(l)

	globalLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(globalLogger)
}

func Logger() *slog.Logger {
	if globalLogger == nil {
		Init("warn")
	}
	return globalLogger
}

func Debug(msg string, args ...any) {
	Logger().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Logger().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Logger().Error(msg, args...)
}
