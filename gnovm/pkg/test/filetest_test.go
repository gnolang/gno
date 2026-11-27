package test

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// A stdlib load on a shared store, as in TestFiles, must not change the gas
// of later filetests.
func TestFiletestGasIndependentOfStoreHistory(t *testing.T) {
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)
	opts := NewTestOptions(rootDir, io.Discard, io.Discard, nil)
	run := func(name, src string) int64 {
		t.Helper()
		_, gas, _, err := opts.RunFiletest(name, []byte(src), opts.TestStore)
		require.NoError(t, err)
		return gas
	}

	// make allocates while preprocessing, so its gas covers preprocessing.
	const measured = `package main

func main() {
	buf := make([]byte, 10)
	_ = string(buf)
}
`
	fresh := run("fresh.gno", measured)
	rootAlloc := opts.TestStore.GetAllocator()
	run("load.gno", "package main\n\nimport \"strings\"\n\nfunc main() { _ = strings.Repeat(\"a\", 3) }\n")
	require.Same(t, rootAlloc, opts.TestStore.GetAllocator(),
		"loading a stdlib must not replace the store's allocator")
	require.Equal(t, fresh, run("after.gno", measured),
		"gas must not depend on what ran earlier on the store")
}
