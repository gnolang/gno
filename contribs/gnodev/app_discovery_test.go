package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/logger"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func TestLogNoWorkspace_Local(t *testing.T) {
	var buf bytes.Buffer
	clogger := logger.NewColumnLogger(&buf, slog.LevelInfo, termenv.Ascii)
	logNoWorkspace(slog.New(clogger).WithGroup(LoaderLogName), localNoWorkspaceHint)

	out := buf.String()
	assert.Contains(t, out, LoaderLogName)
	assert.Contains(t, out, "no workspace")
	assert.Contains(t, out, "gnomod.toml")
	assert.Contains(t, out, "gnowork.toml")
	assert.Contains(t, out, "discovery mode")
	assert.Contains(t, out, "-remote")
	assert.Contains(t, out, "-extra-root")
	// Banner should be visually distinct (multi-line, not a one-liner).
	assert.GreaterOrEqual(t, strings.Count(out, "\n"), 2)
}

func TestLogNoWorkspace_Staging(t *testing.T) {
	var buf bytes.Buffer
	clogger := logger.NewColumnLogger(&buf, slog.LevelInfo, termenv.Ascii)
	logNoWorkspace(slog.New(clogger).WithGroup(LoaderLogName), stagingNoWorkspaceHint)

	out := buf.String()
	assert.Contains(t, out, "no workspace")
	assert.Contains(t, out, "up front")
	assert.Contains(t, out, "-extra-root")
	// Staging eager-loads everything at genesis; nothing resolves lazily.
	assert.NotContains(t, out, "discovery mode")
	assert.NotContains(t, out, "on-demand")
	assert.GreaterOrEqual(t, strings.Count(out, "\n"), 2)
}
