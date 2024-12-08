package main

import (
	"io"
	"log/slog"

	"github.com/charmbracelet/lipgloss"
	"github.com/gnolang/gno/contribs/gnodev/pkg/logger"
	gnolog "github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/muesli/termenv"
)

func setuplogger(cfg *devCfg, out io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if cfg.verbose {
		level = slog.LevelDebug
	}

	if cfg.serverMode {
		zaplogger := logger.NewZapLogger(out, level)
		return gnolog.ZapLoggerToSlog(zaplogger)
	}

	// Detect term color profile
	colorProfile := termenv.DefaultOutput().Profile
	clogger := logger.NewColumnLogger(out, level, colorProfile)

	// Register well known group color with system colors
	clogger.RegisterGroupColor(NodeLogName, lipgloss.Color("3"))
	clogger.RegisterGroupColor(WebLogName, lipgloss.Color("4"))
	clogger.RegisterGroupColor(KeyPressLogName, lipgloss.Color("5"))
	clogger.RegisterGroupColor(EventServerLogName, lipgloss.Color("6"))

	return slog.New(clogger)
}
