package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCmd(t *testing.T) {
	// Initialize stdout listener
	mockOutput := bytes.NewBufferString("")
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOutput))

	// Initialize version command
	versionCmd := newVersionCmd(io)
	originalVersion := version.Version
	t.Cleanup(func() {
		version.Version = originalVersion
	})

	// test: version settled
	versionValue := "chain/test4.2"
	version.Version = versionValue
	require.NoError(t, versionCmd.ParseAndRun(context.Background(), nil))

	output := mockOutput.String()
	expected := "gnodev version: " + versionValue + "\n"
	assert.Equal(
		t,
		expected,
		output,
	)
}
