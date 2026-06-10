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

func TestLogDiscoveryMode_Format(t *testing.T) {
	var buf bytes.Buffer
	clogger := logger.NewColumnLogger(&buf, slog.LevelInfo, termenv.Ascii)
	logDiscoveryMode(slog.New(clogger).WithGroup(LoaderLogName))

	out := buf.String()
	assert.Contains(t, out, LoaderLogName)
	assert.Contains(t, out, "no workspace")
	assert.Contains(t, out, "gnomod.toml")
	assert.Contains(t, out, "gnowork.toml")
	assert.Contains(t, out, "discovery mode")
	assert.Contains(t, out, "-extra-root")
	// Banner should be visually distinct (multi-line, not a one-liner).
	assert.GreaterOrEqual(t, strings.Count(out, "\n"), 2)
}
