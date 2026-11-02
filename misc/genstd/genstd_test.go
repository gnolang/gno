package main

import (
	"flag"
	"os"
	"os/exec"
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

// TestIntegrationInt256Alias generates bindings for natives that take and
// return the 256-bit types under an import alias of their own, and compiles
// the result.
func TestIntegrationInt256Alias(t *testing.T) {
	chdir(t, "testdata/integration_int256")
	skipExternalTools = true

	err := _main(".")
	t.Cleanup(func() { os.Remove("generated.go") })
	require.NoError(t, err)

	got, err := os.ReadFile("generated.go")
	require.NoError(t, err)
	assert.NotContains(t, string(got), "native256")
	assert.Contains(t, string(got), "vm_pkg_gnolang.Uint256")

	out, err := exec.Command("go", "build", ".").CombinedOutput()
	require.NoError(t, err, "generated code does not compile:\n%s", out)
}
