package common

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestLoggerFromServerFlags(t *testing.T) {
	t.Parallel()

	t.Run("invalid log level", func(t *testing.T) {
		t.Parallel()

		// Create the server flags with an invalid log level.
		flags := &ServerFlags{
			LogLevel: "invalid",
		}

		// Create the logger.
		_, _, err := LoggerFromServerFlags(flags, commands.NewTestIO())
		require.Error(t, err)
	})

	t.Run("valid log level", func(t *testing.T) {
		t.Parallel()

		// Create the server flags with a valid log level.
		flags := &ServerFlags{
			LogLevel: zapcore.InfoLevel.String(),
		}

		// Create the logger.
		_, _, err := LoggerFromServerFlags(flags, commands.NewTestIO())
		require.NoError(t, err)
	})
}
