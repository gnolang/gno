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

func TestVersion(t *testing.T) {
	// Not parallel: subtests mutate the global version.Version.
	originalVersion := version.Version
	t.Cleanup(func() {
		version.Version = originalVersion
	})

	versionValues := []string{"chain/test4.2", "develop", "master"}

	for _, v := range versionValues {
		version.Version = v

		t.Run(v, func(t *testing.T) {
			mockOut := bytes.NewBufferString("")
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))

			cmd := newRootCmd(io)
			err := cmd.ParseAndRun(context.Background(), []string{"version"})
			require.NoError(t, err)

			assert.Contains(t, mockOut.String(), "gnoland version: "+v)
		})
	}
}
