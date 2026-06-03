package components

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseImports_ClassifyAndLink(t *testing.T) {
	t.Parallel()
	src := []byte(`package foo
import (
	"strings"
	"gno.land/p/demo/avl"
	"gno.land/r/gnoland/users/v1"
	"github.com/external/dep"
)
func Ignored() {}`)
	got := parseImports(map[string][]byte{"main.gno": src}, "gno.land")
	require.Equal(t, []ImportLink{
		{Path: "github.com/external/dep", Kind: "external", Link: ""},
		{Path: "gno.land/p/demo/avl", Kind: "package", Link: "/p/demo/avl"},
		{Path: "gno.land/r/gnoland/users/v1", Kind: "realm", Link: "/r/gnoland/users/v1"},
		{Path: "strings", Kind: "stdlib", Link: ""},
	}, got)
}

func TestParseImports_DedupAcrossFiles(t *testing.T) {
	t.Parallel()
	src1 := []byte(`package p
import "strings"
import "gno.land/p/demo/avl"`)
	src2 := []byte(`package p
import "strings"
import "fmt"`)
	got := parseImports(map[string][]byte{"a.gno": src1, "b.gno": src2}, "gno.land")
	paths := make([]string, 0, len(got))
	for _, im := range got {
		paths = append(paths, im.Path)
	}
	require.Equal(t, []string{"fmt", "gno.land/p/demo/avl", "strings"}, paths)
}

func TestParseImports_MalformedBodyTolerated(t *testing.T) {
	t.Parallel()
	// ImportsOnly stops after imports, so .gno-only syntax later is irrelevant.
	src := []byte(`package foo
import "strings"

func WithCross(cur realm, arg string) string {
	return arg
}`)
	got := parseImports(map[string][]byte{"main.gno": src}, "gno.land")
	require.Len(t, got, 1)
	require.Equal(t, "strings", got[0].Path)
}

func TestParseImports_EmptyInput(t *testing.T) {
	t.Parallel()
	got := parseImports(nil, "gno.land")
	require.Nil(t, got)
}

func TestParseImports_UnparseableFileSilentlySkipped(t *testing.T) {
	t.Parallel()
	src := []byte(`not go at all`)
	got := parseImports(map[string][]byte{"bad.gno": src}, "gno.land")
	require.Nil(t, got)
}
