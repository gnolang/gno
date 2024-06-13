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

func Test_Scripts(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	require.NoError(t, err)

	testdirs, err := os.ReadDir(testdata)
	require.NoError(t, err)

	for _, dir := range testdirs {
		if !dir.IsDir() {
			continue
		}

		name := dir.Name()
		t.Logf("testing: %s", name)
		t.Run(name, func(t *testing.T) {
			updateScripts, _ := strconv.ParseBool(os.Getenv("UPDATE_SCRIPTS"))
			p := testscript.Params{
				UpdateScripts: updateScripts,
				Dir:           filepath.Join(testdata, name),
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
