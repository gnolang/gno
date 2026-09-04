package integration

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

// The adapter stands in for testing.T, whose Log formats like Println: a space
// between every operand, including two strings.
func TestAdapterLogSeparatesOperands(t *testing.T) {
	var buf bytes.Buffer
	adapter := NewTestscriptT(slog.New(slog.NewTextHandler(&buf, nil)), false)

	adapter.Run("tour", func(testscript.T) {})

	require.Contains(t, buf.String(), "=== RUN tour")
	require.Contains(t, buf.String(), "--- PASS tour")
}
