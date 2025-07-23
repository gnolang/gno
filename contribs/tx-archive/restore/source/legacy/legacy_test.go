package legacy

import (
	"context"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
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

func TestSource_Legacy(t *testing.T) {
	t.Parallel()

	t.Run("no source found", func(t *testing.T) {
		t.Parallel()

		source, err := NewSource("./dummy-file.txt")
		require.Nil(t, source)
		require.Error(t, err)
	})

	t.Run("invalid parsing", func(t *testing.T) {
		t.Parallel()

		// Create a temp file
		tempFile := createTempFile(t)

		// Temp file cleanup
		t.Cleanup(func() {
			require.NoError(t, tempFile.Close())
			require.NoError(t, os.Remove(tempFile.Name()))
		})

		// Write invalid JSON to file
		_, err := tempFile.WriteString(`{"example": 123`) // invalid JSON
		require.NoError(t, err)

		// Create the standard source
		source, err := NewSource(tempFile.Name())
		require.NoError(t, err)

		// Try to parse the file
		nextTx, nextErr := source.Next(context.Background())
		require.Nil(t, nextTx)
		require.Error(t, nextErr)
	})

	t.Run("valid parsing", func(t *testing.T) {
		t.Parallel()

		// Create a temp file
		tempFile := createTempFile(t)

		// Temp file cleanup
		t.Cleanup(func() {
			require.NoError(t, tempFile.Close())
			require.NoError(t, os.Remove(tempFile.Name()))
		})

		// Write a standard format to the temp file
		tx := std.Tx{
			Memo: "example tx",
		}

		txRaw, err := amino.MarshalJSON(tx)
		require.NoError(t, err)

		_, err = tempFile.Write(txRaw)
		require.NoError(t, err)

		// Create the legacy source
		source, err := NewSource(tempFile.Name())
		require.NoError(t, err)

		// Try to parse the file
		nextTx, nextErr := source.Next(context.Background())
		require.NoError(t, nextErr)
		require.NotNil(t, nextTx)

		assert.Equal(t, tx, *nextTx)
	})
}
