package log_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestDummy(t *testing.T) {
	l, err := log.NewTMLogger(os.Stdout, slog.LevelDebug)
	require.NoError(t, err)

	l.Info("info!", "message", 123)
	l.Debug("debug!", "message", 123)
	l.Error("error!", "message", 123)
	l.Warn("warn!", "message", 123)
}
