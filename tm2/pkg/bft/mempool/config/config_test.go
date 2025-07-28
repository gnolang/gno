package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()

		assert.NoError(t, cfg.ValidateBasic())
	})

	t.Run("invalid max tx count", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.MaxTxCount = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidMaxTxCount)
	})

	t.Run("invalid max pool size", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.MaxBytes = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidMaxBytes)
	})
}
