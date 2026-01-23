//go:build gnobench

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func Test_ScriptsBench(t *testing.T) {
	testdir, err := filepath.Abs(filepath.Join("testdata", "bench"))
	require.NoError(t, err)

	// Skip if bench directory doesn't exist
	if _, err := os.Stat(testdir); os.IsNotExist(err) {
		t.Skip("testdata/bench directory not found")
	}

	homeDir, buildDir := t.TempDir(), t.TempDir()
	p := integration.NewTestingParams(t, testdir)

	if coverdir, ok := integration.ResolveCoverageDir(); ok {
		err := integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	err = integration.SetupGnoBench(&p, homeDir, buildDir)
	require.NoError(t, err)

	testscript.Run(t, p)
}
