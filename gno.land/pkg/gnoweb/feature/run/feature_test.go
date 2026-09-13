package run

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func validDeps() Deps {
	return Deps{
		Logger:  discardLogger(),
		Domain:  "gno.land",
		Remote:  "http://localhost:26657",
		ChainId: "test",
	}
}

func TestNewPanics(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		setup   func(*Deps)
		wantMsg string
	}{
		{
			name:    "empty remote",
			setup:   func(deps *Deps) { deps.Remote = "" },
			wantMsg: "run.New: Remote RPC endpoint is required",
		},
		{
			name:    "empty chain id",
			setup:   func(deps *Deps) { deps.ChainId = "" },
			wantMsg: "run.New: Chain ID is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := validDeps()
			tc.setup(&deps)

			assert.PanicsWithValue(t, tc.wantMsg, func() {
				New(deps)
			})
		})
	}
}

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	t.Run("defaults domain and logger", func(t *testing.T) {
		t.Parallel()

		deps := validDeps()
		deps.Domain = ""
		deps.Logger = nil

		h := New(deps)
		require.NotNil(t, h)
		assert.Equal(t, "gno.land", h.deps.Domain)
		assert.NotNil(t, h.deps.Logger)
	})

	t.Run("preserves explicit domain and logger", func(t *testing.T) {
		t.Parallel()
		logger := discardLogger()
		deps := validDeps()
		deps.Domain = "example.com"
		deps.Logger = logger

		h := New(deps)
		require.NotNil(t, h)
		assert.Equal(t, "example.com", h.deps.Domain)
		assert.Same(t, logger, h.deps.Logger)
	})
}
