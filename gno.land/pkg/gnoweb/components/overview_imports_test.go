package components

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildImports_ClassifyAndLink(t *testing.T) {
	t.Parallel()
	// Input is the sorted, deduplicated path list as produced by vm/qdoc.
	got := buildImports([]string{
		"github.com/external/dep",
		"gno.land/p/demo/avl",
		"gno.land/r/gnoland/users/v1",
		"strings",
	}, "gno.land")
	require.Equal(t, []ImportLink{
		{Path: "github.com/external/dep", Kind: "external", Link: ""},
		{Path: "gno.land/p/demo/avl", Kind: "package", Link: "/p/demo/avl"},
		{Path: "gno.land/r/gnoland/users/v1", Kind: "realm", Link: "/r/gnoland/users/v1"},
		{Path: "strings", Kind: "stdlib", Link: stdlibSourceBase + "strings"},
	}, got)
}

func TestBuildImports_StdlibLinksUpstream(t *testing.T) {
	t.Parallel()
	// Stdlibs ship with the node rather than being deployed, so they have no
	// package page and link to their upstream source instead.
	got := buildImports([]string{"chain/banker", "errors"}, "gno.land")
	require.Equal(t, []ImportLink{
		{Path: "chain/banker", Kind: "stdlib", Link: stdlibSourceBase + "chain/banker"},
		{Path: "errors", Kind: "stdlib", Link: stdlibSourceBase + "errors"},
	}, got)
}

func TestBuildImports_Empty(t *testing.T) {
	t.Parallel()
	require.Nil(t, buildImports(nil, "gno.land"))
	require.Nil(t, buildImports([]string{}, "gno.land"))
}
