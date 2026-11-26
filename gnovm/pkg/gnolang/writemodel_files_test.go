package gnolang_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/stretchr/testify/require"
)

// Part B: every zrealm file test runs with the model beside the VM at every
// call start (writemodel_vm_test.go); the file tests' own outcomes are not judged.
func TestWriteModelFiles(t *testing.T) {
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)
	opts := &test.TestOptions{RootDir: rootDir, Output: io.Discard, Error: io.Discard}
	opts.BaseStore, opts.TestStore = test.StoreWithOptions(rootDir, opts.WriterForStore(),
		test.StoreOptions{WithExtern: true, WithExamples: true, Testing: true})

	dir := "../../tests/files"
	paths, err := filepath.Glob(filepath.Join(dir, "zrealm_*.gno"))
	require.NoError(t, err)
	setFile, stop := gnolang.WMObserve()
	defer stop()
	files, failed := 0, 0
	for _, p := range paths {
		name := filepath.Base(p)
		if strings.HasSuffix(name, "_long.gno") && testing.Short() {
			continue
		}
		content, err := os.ReadFile(p)
		require.NoError(t, err)
		setFile(name)
		files++
		if _, _, _, err := opts.RunFiletest(name, content, opts.TestStore); err != nil {
			failed++
		}
	}
	c := stop()
	bad := c.Mismatches + c.Drift + c.Unaligned + c.Broken
	t.Logf("MODEL: %d files, %d calls, %d mismatches", files, c.Calls, bad)
	t.Logf("MODEL-TAKEN: 0 calls took loan liveness from the VM (deferred); %d deferred calls read their arguments", c.Deferred)
	t.Logf("MODEL-DETAIL: %d answers differ, %d caller modes drifted, %d stacks unaligned, %d glue failures",
		c.Mismatches, c.Drift, c.Unaligned, c.Broken)
	t.Logf("MODEL-COVER: %d files failed their own check; answers by kind %v; %d loans made; %d calls ran lent", failed, c.Kinds, c.Loans, c.LentHits)
	for _, s := range c.Samples {
		t.Log(s)
	}
	if bad != 0 || c.Calls == 0 {
		t.Errorf("model and VM differ")
	}
}
