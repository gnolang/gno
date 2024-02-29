package main

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/integration"
)

func Test_ScriptsTranspile(t *testing.T) {
	p := testscript.Params{
		Dir: "testdata/gno_transpile",
	}

	if coverdir, ok := integration.ResolveCoverageDir(); ok {
		err := integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	err := integration.SetupGno(&p, t.TempDir())
	require.NoError(t, err)

	testscript.Run(t, p)
}
