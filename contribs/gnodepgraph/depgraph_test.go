package main

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleRequire(t *testing.T) {
	allPkgs, err := gnomod.ListPkgs("../../examples")
	require.NoError(t, err)

	for _, pkg := range allPkgs {
		if pkg.Name == "gno.land/r/gnoland/blog" {
			visited := make(map[string]bool)
			graphData := ""

			err = buildGraphData(pkg, allPkgs, visited, make(map[string]bool), &graphData)
			require.NoError(t, err)

			assert.Contains(t, graphData, "\"gno.land/r/gnoland/blog\" -> \"gno.land/p/demo/blog\"")
		}
	}
}

// test for big graph

// test for fail on cyclical dependencies

// test for fail on missing dependencies

// test for not duplicating dependencies
