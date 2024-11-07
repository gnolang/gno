package main

import (
	"testing"

	gnoland_integration "github.com/gnolang/gno/gno.land/pkg/integration"
	gno_integration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/stretchr/testify/require"
)

func TestTestdata(t *testing.T) {
	p := gno_integration.NewTestingParams(t, "testdata")

	if coverdir, ok := gno_integration.ResolveCoverageDir(); ok {
		err := gno_integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	// Set up gnoland for testscript
	err := gnoland_integration.SetupGnolandTestscript(t, &p)
	require.NoError(t, err)

	// Run testscript
	// XXX: We have to use seqshim for now as tests don't run well in parallel
	gnoland_integration.RunSeqShimTestscripts(t, p)
}
