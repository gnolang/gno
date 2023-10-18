package main

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func Test_ScriptsBuild(t *testing.T) {
	p := testscript.Params{
		Dir: "testdata/gno_build",
	}

	err := integration.SetupCoverage(&p)
	require.NoError(t, err)
	err = integration.SetupGno(&p, t.TempDir())
	require.NoError(t, err)

	testscript.Run(t, p)
}
