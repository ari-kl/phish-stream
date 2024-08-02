package util

import (
	"log/slog"
	"os"

	"github.com/ari-kl/phish-stream/config"
)

var Logger = slog.Default()

var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func SetupLogger() {
	var logLevel slog.Level

	if level, ok := logLevels[config.LogLevel]; ok {
		logLevel = level
	} else {
		// We don't need to wait for the proper level to be set before logging this
		// The default level is info, so we can log this message with the default logger
		Logger.Warn("Invalid log level, defaulting to info")
		logLevel = slog.LevelInfo
	}

	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

}
