package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func Test_Scripts(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	require.NoError(t, err)

	testdirs, err := os.ReadDir(testdata)
	require.NoError(t, err)

	homeDir, buildDir := t.TempDir(), t.TempDir()
	for _, dir := range testdirs {
		if !dir.IsDir() {
			continue
		}

		name := dir.Name()
		t.Run(name, func(t *testing.T) {
			testdir := filepath.Join(testdata, name)
			p := integration.NewTestingParams(t, testdir)
			if coverdir, ok := integration.ResolveCoverageDir(); ok {
				err := integration.SetupTestscriptsCoverage(&p, coverdir)
				require.NoError(t, err)
			}

			err := integration.SetupGno(&p, homeDir, buildDir)
			require.NoError(t, err)

			testscript.Run(t, p)
		})
	}
}
