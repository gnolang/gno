package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	assert.Equal(t, string(want), string(got))
}
