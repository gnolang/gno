package backup

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("invalid output file", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.OutputFile = "" // invalid output file location

		assert.ErrorIs(t, ValidateConfig(cfg), errInvalidOutputLocation)
	})

	t.Run("output file already exists, no overwrite", func(t *testing.T) {
		t.Parallel()

		// Create temp file
		tempFile := createTempFile(t)

		t.Cleanup(func() {
			require.NoError(t, tempFile.Close())
			require.NoError(t, os.Remove(tempFile.Name()))
		})

		cfg := DefaultConfig()
		cfg.OutputFile = tempFile.Name()
		cfg.Overwrite = false

		assert.ErrorIs(t, ValidateConfig(cfg), errOutputFileExists)
	})

	t.Run("output file already exists, with overwrite", func(t *testing.T) {
		t.Parallel()

		// Create temp file
		tempFile := createTempFile(t)

		t.Cleanup(func() {
			require.NoError(t, tempFile.Close())
			require.NoError(t, os.Remove(tempFile.Name()))
		})

		cfg := DefaultConfig()
		cfg.OutputFile = tempFile.Name()
		cfg.Overwrite = true

		assert.NoError(t, ValidateConfig(cfg))
	})

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

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()

		assert.NoError(t, ValidateConfig(cfg))
	})
}
