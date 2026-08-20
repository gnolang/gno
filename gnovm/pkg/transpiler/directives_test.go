package transpiler

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The generated file is Go, so a directive inherited from the Gno source stops
// being inert and starts steering the Go toolchain.
func TestTranspileStripsInheritedDirectives(t *testing.T) {
	t.Parallel()

	const source = `//go:build ignore

package tr

//go:generate echo PWNED

//line forged.gno:999:1

//nolint:gosec
// A normal comment.
func F() int { return 1 }
`
	res, err := Transpile(source, "gno", "tr.gno")
	require.NoError(t, err)
	out := res.Translated

	for _, gone := range []string{
		"//go:build ignore",       // would collide with the header's //go:build
		"//go:generate echo",      // would run under `go generate`
		"//line forged.gno:999:1", // would forge compiler positions
		"//nolint:gosec",          // would suppress findings for whoever lints it
	} {
		assert.NotContains(t, out, gone, "inherited directive must not reach the output")
	}

	// What the transpiler writes itself, and ordinary content, must survive.
	assert.Contains(t, out, "//go:build gno")
	assert.Contains(t, out, "//line tr.gno:1:1")
	assert.Contains(t, out, "// A normal comment.")
	assert.Contains(t, out, "func F() int { return 1 }")

	// Exactly one //go:build: the collision is what made `go build` reject the
	// generated file with "multiple //go:build comments".
	assert.Equal(t, 1, strings.Count(out, "//go:build"))
}

func TestIsNolintComment(t *testing.T) {
	t.Parallel()

	for text, want := range map[string]bool{
		"//nolint":              true,
		"//nolint:gosec":        true,
		"//nolint:gosec // why": true,
		"//nolint // why":       true,
		"//nolintfoo:bar":       false,
		"// nolint:gosec":       false, // a space makes it prose to the linter
		"//nolinting is fun":    false,
	} {
		assert.Equal(t, want, isNolintComment(text), "%q", text)
	}
}

// Blanking rather than deleting is what keeps the "//line" header honest: the
// generated file must have one line per source line, or every position after a
// removed comment points somewhere earlier. Deleting them reported an error
// truly on line 7 as line 5.
func TestTranspilePreservesLineCount(t *testing.T) {
	t.Parallel()

	const source = `package tr

//go:noinline

//nolint:gosec
/*line forged.gno:99:1*/
// doc
func F() int {
	return 1
}
`
	res, err := Transpile(source, "gno", "tr.gno")
	require.NoError(t, err)

	// Drop the header the transpiler writes, then compare what is left with the
	// source line for line.
	_, body, ok := strings.Cut(res.Translated, "//line tr.gno:1:1\n")
	require.True(t, ok, "expected the //line header")
	assert.Equal(t, strings.Count(source, "\n"), strings.Count(body, "\n"),
		"generated body must keep one line per source line")
}

// `go generate` scans physical lines and never parses, so a "//go:generate"
// line inside a block comment runs even though Go treats it as commentary.
func TestTranspileNeutralizesGenerateInBlockComment(t *testing.T) {
	t.Parallel()

	const source = `package tr

/*
//go:generate echo PWNED
*/
func F() int { return 1 }
`
	res, err := Transpile(source, "gno", "tr.gno")
	require.NoError(t, err)
	assert.NotContains(t, res.Translated, "//go:generate",
		"a directive line inside a block comment must not survive")
	_, body, ok := strings.Cut(res.Translated, "//line tr.gno:1:1\n")
	require.True(t, ok)
	assert.Equal(t, strings.Count(source, "\n"), strings.Count(body, "\n"),
		"neutralizing inside a block must keep the block's line count")
}

// Blanking keeps every CommentGroup non-empty. An emptied group is invalid --
// ast.CommentGroup.Pos() indexes List[0] -- and Result.File hands the tree to
// callers, Doc fields included.
func TestTranspileLeavesNoEmptyCommentGroup(t *testing.T) {
	t.Parallel()

	const source = `package tr

//go:noinline
func F() int { return 1 }
`
	res, err := Transpile(source, "gno", "tr.gno")
	require.NoError(t, err)

	for _, cg := range res.File.Comments {
		require.NotEmpty(t, cg.List, "an empty CommentGroup is invalid")
		assert.NotPanics(t, func() { _, _ = cg.Pos(), cg.End() })
	}
	for _, d := range res.File.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Doc == nil {
			continue
		}
		require.NotEmpty(t, fd.Doc.List, "a Doc group must not be left empty")
		assert.NotPanics(t, func() { _, _ = fd.Doc.Pos(), fd.Doc.End() })
	}
}
