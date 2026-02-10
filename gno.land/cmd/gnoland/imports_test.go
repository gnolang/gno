package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoTestingStdlibImport(t *testing.T) {
	t.Skip("as keeper is now also typchecking `_test` files, this test doesn't make sense anymore")

	// See: https://github.com/gnolang/gno/issues/3585
	// The gno.land binary should not import testing stdlibs, which contain unsafe
	// code in the respective native bindings.

	res, err := exec.Command("go", "list", "-f", `{{ join .Deps "\n" }}`, ".").CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(res), "github.com/gnolang/gno/gnovm/stdlibs\n", "should contain normal stdlibs")
	assert.NotContains(t, string(res), "github.com/gnolang/gno/gnovm/tests/stdlibs\n", "should not contain test stdlibs")
}
