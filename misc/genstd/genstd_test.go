package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update-golden-tests", false, "update golden test files")

func TestIntegration(t *testing.T) {
	chdir(t, "testdata/integration")
	skipExternalTools = true

	err := _main(".")
	t.Cleanup(func() { os.Remove("generated.go") })
	require.NoError(t, err)

	got, err := os.ReadFile("generated.go")
	require.NoError(t, err)

	want, err := os.ReadFile("generated.go.golden")
	require.NoError(t, err)

	if *update {
		require.NoError(t, os.WriteFile("generated.go.golden", got, 0o644))
	} else {
		assert.Equal(t, string(want), string(got))
	}
}
