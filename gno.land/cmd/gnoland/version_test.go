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
	t.Parallel()

	originalVersion := version.Version
	t.Cleanup(func() {
		version.Version = originalVersion
	})

	versionValues := []string{"develop", "master.42+abc1234", "v1.0.0"}

	for _, v := range versionValues {
		t.Run(v, func(t *testing.T) {
			version.Version = v

			mockOut := bytes.NewBufferString("")
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))

			cmd := newVersionCmd(io)
			require.NoError(t, cmd.ParseAndRun(context.Background(), nil))

			assert.Equal(t, "gnoland version: "+v+"\n", mockOut.String())
		})
	}
}
