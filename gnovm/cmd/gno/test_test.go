package main

import (
	"os"
	"strconv"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func Test_ScriptsTest(t *testing.T) {
	updateScripts, _ := strconv.ParseBool(os.Getenv("UPDATE_SCRIPTS"))
	p := testscript.Params{
		UpdateScripts: updateScripts,
		Dir:           "testdata/gno_test",
	}

	if coverdir, ok := integration.ResolveCoverageDir(); ok {
		err := integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	err := integration.SetupGno(&p, t.TempDir())
	require.NoError(t, err)

	testscript.Run(t, p)
}
