package main

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/charmbracelet/lipgloss"
	"github.com/gnolang/gno/contribs/gnodev/pkg/logger"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/muesli/termenv"
	"go.uber.org/zap/zapcore"
)

func setuplogger(cfg *AppConfig, out io.Writer) (*slog.Logger, error) {
	level := slog.LevelInfo
	if cfg.verbose {
		level = slog.LevelDebug
	}

	// Set up the logger
	switch cfg.logFormat {
	case "json":
		return newJSONLogger(out, level), nil
	case "console", "":
		// Detect term color profile
		colorProfile := termenv.DefaultOutput().Profile

		clogger := logger.NewColumnLogger(out, level, colorProfile)

		// Register well known group color with system colors
		clogger.RegisterGroupColor(NodeLogName, lipgloss.Color("3"))
		clogger.RegisterGroupColor(WebLogName, lipgloss.Color("4"))
		clogger.RegisterGroupColor(KeyPressLogName, lipgloss.Color("5"))
		clogger.RegisterGroupColor(EventServerLogName, lipgloss.Color("6"))

		return slog.New(clogger), nil
	default:
		return nil, fmt.Errorf("invalid log format %q", cfg.logFormat)
	}
}

func newJSONLogger(w io.Writer, level slog.Level) *slog.Logger {
	var zaplevel zapcore.Level
	switch level {
	case slog.LevelDebug:
		zaplevel = zapcore.DebugLevel
	case slog.LevelInfo:
		zaplevel = zapcore.InfoLevel
	case slog.LevelWarn:
		zaplevel = zapcore.WarnLevel
	case slog.LevelError:
		zaplevel = zapcore.ErrorLevel
	default:
		panic("unknown slog level: " + level.String())
	}

	return log.ZapLoggerToSlog(log.NewZapJSONLogger(w, zaplevel))
}
