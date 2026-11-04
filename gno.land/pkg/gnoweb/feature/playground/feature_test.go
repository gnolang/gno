package playground

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validDeps() Deps {
	return Deps{
		Client:   &stubClient{},
		Packages: &stubPackages{},
		Logger:   discardLogger(),
		Domain:   "gno.land",
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
			name:    "nil client",
			setup:   func(deps *Deps) { deps.Client = nil },
			wantMsg: "playground.New: Client is required",
		},
		{
			name:    "nil packages",
			setup:   func(deps *Deps) { deps.Packages = nil },
			wantMsg: "playground.New: Packages is required",
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
