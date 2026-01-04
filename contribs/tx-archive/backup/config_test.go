package backup

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createTempFile creates a temporary file
func createTempFile(t *testing.T) *os.File {
	t.Helper()

	f, err := os.CreateTemp("", "temp-")
	if err != nil {
		t.Fatalf("unable to create temporary file, %v", err)
	}

	return f
}

func TestConfig_ValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid block range", func(t *testing.T) {
		t.Parallel()

		var (
			fromBlock uint64 = 10
			toBlock          = fromBlock - 1 // to < from
		)

		cfg := DefaultConfig()
		cfg.FromBlock = fromBlock
		cfg.ToBlock = &toBlock

		assert.ErrorIs(t, ValidateConfig(cfg), errInvalidRange)
	})

	t.Run("invalid from block", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.FromBlock = 0 // genesis

		assert.ErrorIs(t, ValidateConfig(cfg), errInvalidFromBlock)
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()

		assert.NoError(t, ValidateConfig(cfg))
	})
}
