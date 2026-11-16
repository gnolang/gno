package config

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("invalid min gas prices", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name   string
			prices string
		}{
			{"invalid gas", "10token/1"},
			{"invalid min gas prices invalid gas denom", "9token/0gs"},
			{"invalid min gas prices zero gas", "10token/0gas"},
			{"invalid min gas prices no gas", "10token/gas"},
			{"invalid min gas prices negtive gas", "10token/-1gas"},
			{"invalid min gas prices invalid denom", "10$token/2gas"},
			{"invalid min gas prices invalid second denom", "10token/2gas;10/3gas"},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				cfg := DefaultAppConfig()
				cfg.MinGasPrices = testCase.prices

				assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidMinGasPrices)
			})
		}
	})

	t.Run("valid min gas prices", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultAppConfig()
		cfg.MinGasPrices = "10foo/3gas;5bar/3gas"

		assert.NoError(t, cfg.ValidateBasic())
	})

	t.Run("invalid prune strategy", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultAppConfig()
		cfg.PruneStrategy = "best-effort"

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidPruneStrategy)
	})

	t.Run("valid prune strategy", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultAppConfig()
		cfg.PruneStrategy = types.PruneEverythingStrategy

		assert.NoError(t, cfg.ValidateBasic())
	})

	t.Run("negative query limits", func(t *testing.T) {
		t.Parallel()

		for name, set := range map[string]func(*AppConfig){
			"gas":     func(c *AppConfig) { c.QueryMaxGas = -1 },
			"alloc":   func(c *AppConfig) { c.QueryMaxAlloc = -1 },
			"timeout": func(c *AppConfig) { c.QueryTimeout = -1 },
		} {
			cfg := DefaultAppConfig()
			set(cfg)
			assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidQueryLimit, name)
		}
	})

	t.Run("valid query limits", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultAppConfig()
		cfg.QueryMaxGas = 1
		cfg.QueryMaxAlloc = 1
		cfg.QueryTimeout = time.Millisecond

		assert.NoError(t, cfg.ValidateBasic())
	})

	t.Run("valid default config", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultAppConfig()

		assert.NoError(t, cfg.ValidateBasic())
	})
}
