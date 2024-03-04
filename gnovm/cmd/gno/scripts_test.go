package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func TestScript(t *testing.T) {
	updateScripts, _ := strconv.ParseBool(os.Getenv("UPDATE_SCRIPTS"))
	dirs, err := os.ReadDir("testdata")
	require.NoError(t, err)
	for _, dir := range dirs {
		t.Run(dir.Name(), func(t *testing.T) {
			p := testscript.Params{
				UpdateScripts: updateScripts,
				Dir:           filepath.Join("testdata", dir.Name()),
			}

			if coverdir, ok := integration.ResolveCoverageDir(); ok {
				err := integration.SetupTestscriptsCoverage(&p, coverdir)
				require.NoError(t, err)
			}

			err := integration.SetupGno(&p, t.TempDir())
			require.NoError(t, err)

			testscript.Run(t, p)
		})
	}
}
