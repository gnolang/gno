package client

import (
	"bytes"
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Version(t *testing.T) {
	t.Parallel()

	// Initialize stdout listener
	mockOutput := bytes.NewBufferString("")
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOutput))

	// Initialize version command
	versionCmd := NewVersionCmd(nil, io)
	versionValues := []string{"develop", "chain/test4.2"}

	// test: original version
	require.NoError(t, versionCmd.ParseAndRun(context.Background(), nil))

	output := mockOutput.String()
	expected := "gnokey version: " + versionValues[0] + "\n"
	assert.Equal(
		t,
		expected,
		output,
	)

	mockOutput.Reset()

	// test: version settled
	version.Version = versionValues[1]
	require.NoError(t, versionCmd.ParseAndRun(context.Background(), nil))

	output = mockOutput.String()
	expected = "gnokey version: " + versionValues[1] + "\n"
	assert.Equal(
		t,
		expected,
		output,
	)
}
