package restore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid remote address", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Remote = ""

		assert.ErrorIs(t, ValidateConfig(cfg), errInvalidRemote)
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()

		assert.NoError(t, ValidateConfig(cfg))
	})
}
