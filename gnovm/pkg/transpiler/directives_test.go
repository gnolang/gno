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
		"// nolint:gosec":       true, // golangci-lint trims "/ " first, so this suppresses too
		"//  nolint:gosec":      true,
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

	// The import block matters: format.Node re-parses when a file has one
	// (hasUnsortedImports), which is the path where an emptied doc comment is
	// dropped outright rather than left as a blank line. Without it this test
	// passed while a directive in doc position still cost a line.
	const source = `package tr

import (
	"errors"
)

//go:noinline
func F() error {
	return errors.New("x")
}

//nolint:gosec
/*line forged.gno:99:1*/
// doc
func G() int {
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

// go/printer.stripCommonPrefix drops the common leading whitespace of a block
// comment's interior when printing, so an indented directive in the source
// arrives at column 1 in the output -- which is exactly where `go generate`
// looks. Matching the source line rather than the printed one let these
// through: of 2688 block shapes probed in review, 76 leaked.
func TestTranspileNeutralizesIndentedGenerateInBlock(t *testing.T) {
	t.Parallel()

	for name, source := range map[string]string{
		"one space":  "package tr\n\n/*\n //go:generate echo X\n*/\nfunc F() {}\n",
		"two spaces": "package tr\n\n/*\n  //go:generate echo X\n*/\nfunc F() {}\n",
		"tab":        "package tr\n\n/*\n\t//go:generate echo X\n*/\nfunc F() {}\n",
		"star style": "package tr\n\n/*\n * //go:generate echo X\n */\nfunc F() {}\n",
		// The opener does not shelter a directive on its own line:
		// formatDocComment moves "/*" onto a line of its own, leaving the
		// directive at column 1.
		"on the opener line": "package tr\n\n/*//go:generate echo X\nprose\n*/\nfunc F() {}\n",
		"opener, indented":   "package tr\n\n/*  //go:generate echo X\nprose\n*/\nfunc F() {}\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			res, err := Transpile(source, "gno", "tr.gno")
			require.NoError(t, err)
			for _, line := range strings.Split(res.Translated, "\n") {
				assert.False(t, strings.HasPrefix(line, "//go:generate"),
					"a directive must not reach column 1 of the output: %q", line)
			}
		})
	}
}

// Blanking a line that carries the terminator must keep it, or Result.File
// holds a comment that never closes.
func TestTranspileKeepsBlockCommentTerminated(t *testing.T) {
	t.Parallel()

	res, err := Transpile("package tr\n\n/*\n//go:generate echo X*/\nfunc F() {}\n", "gno", "tr.gno")
	require.NoError(t, err)
	for _, cg := range res.File.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(c.Text, "/*") {
				assert.True(t, strings.HasSuffix(c.Text, "*/"),
					"block comment must stay terminated, got %q", c.Text)
			}
		}
	}
}

// The whitespace rule was approximated wrong twice -- first missing tabs, then a
// vertical tab -- each time letting an indented directive reach column 1 of the
// output, because go/printer strips any prefix byte <= \' \' (or \'*\') from a
// block comment's interior. So assert the property over every byte the printer
// could strip, rather than the few spellings already known to fail.
func TestNoDirectiveReachesColumnOne(t *testing.T) {
	t.Parallel()

	for b := 1; b <= 0x20; b++ {
		if b == '\n' || b == '\r' {
			continue // ends the line rather than indenting it
		}
		for _, prefix := range []string{
			string(rune(b)),
			string(rune(b)) + string(rune(b)),
			"*" + string(rune(b)),
			string(rune(b)) + "*" + string(rune(b)),
		} {
			source := "package tr\n\n/*\n" + prefix + "//go:generate echo X\n*/\nfunc F() {}\n"
			res, err := Transpile(source, "gno", "tr.gno")
			if err != nil {
				continue // not every byte makes a parseable file
			}
			for _, line := range strings.Split(res.Translated, "\n") {
				require.False(t, strings.HasPrefix(line, "//go:generate"),
					"byte %#x with prefix %q let a directive reach column 1", b, prefix)
			}
		}
	}
}

// Go's directive classifier excludes the legacy "// +build" form, but
// go/printer synthesizes a matching "//go:build" for it, which then collides
// with the header the transpiler writes.
func TestTranspileNeutralizesLegacyBuildConstraint(t *testing.T) {
	t.Parallel()

	res, err := Transpile("// +build ignore\n\npackage tr\n\nfunc F() {}\n", "gno", "tr.gno")
	require.NoError(t, err)
	assert.NotContains(t, res.Translated, "+build ignore")
	assert.Equal(t, 1, strings.Count(res.Translated, "//go:build"),
		"a synthesized //go:build would collide with the header and break `go build`")
}

// The line-count invariant, asserted over shapes rather than examples. Two
// separate bugs lived here: an emptied line comment vanished in doc position,
// and an emptied block-comment interior collapsed the block. Both only appear
// with a parenthesized import block, which is what makes format.Node re-parse
// and route the comment through formatDocComment.
func TestLineParityAcrossShapes(t *testing.T) {
	t.Parallel()

	indents := []string{"", " ", "  ", "\t", " * ", "*", "\v", "\f"}
	bodies := []string{"//go:generate echo X", "//go:build ignore", "//line f.gno:9:1", "ordinary text"}
	positions := []string{"doc", "floating", "inline"}
	// The import block is load-bearing: without one the bugs are invisible.
	imports := []string{"", "import (\n\t\"errors\"\n)\n\n"}

	for _, imp := range imports {
		for _, pos := range positions {
			for _, ind := range indents {
				for _, b := range bodies {
					block := "/*\n" + ind + b + "\n*/\n"
					var src string
					switch pos {
					case "doc":
						src = "package tr\n\n" + imp + block + "func F() {}\n"
					case "floating":
						src = "package tr\n\n" + imp + block + "\nfunc F() {}\n"
					case "inline":
						src = "package tr\n\n" + imp + "func F() {\n" + block + "}\n"
					}
					res, err := Transpile(src, "gno", "tr.gno")
					if err != nil {
						continue
					}
					_, out, ok := strings.Cut(res.Translated, "//line tr.gno:1:1\n")
					if !ok {
						continue
					}
					require.Equal(t, strings.Count(src, "\n"), strings.Count(out, "\n"),
						"line drift: position=%s indent=%q body=%q", pos, ind, b)
				}
			}
		}
	}
}

// A directive sharing a doc group with prose must come out exactly as an
// already-inert directive would. go/printer moves directives after the prose
// and inserts a separator -- but it does that to the original too, which is why
// the marker is directive-shaped: the line keeps its classification, so the
// printer treats it exactly as it treated what it replaced.
//
// Asserted as equivalence rather than against the source line count, because
// transpiling legitimately changes that on its own (import rewriting).
func TestMixedDocGroupMatchesInertEquivalent(t *testing.T) {
	t.Parallel()

	for name, doc := range map[string]string{
		"directive then prose":    "//go:noinline\n// prose doc",
		"prose then directive":    "// prose doc\n//go:noinline",
		"prose, directive, prose": "// one\n//go:noinline\n// two",
		"two directives":          "//go:noinline\n//go:nosplit\n// prose",
		"nolint then prose":       "//nolint:gosec\n// prose doc",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			build := func(d string) string {
				return "package tr\n\nimport (\n\t\"errors\"\n)\n\n" + d +
					"\nfunc F() error { return errors.New(\"x\") }\n"
			}
			got, err := Transpile(build(doc), "gno", "tr.gno")
			require.NoError(t, err)

			// The same source with every directive already neutralized: the
			// transpiler must produce byte-identical output.
			inert := doc
			for _, d := range []string{"//go:noinline", "//go:nosplit", "//nolint:gosec"} {
				inert = strings.ReplaceAll(inert, d, neutralizedMarker)
			}
			want, err := Transpile(build(inert), "gno", "tr.gno")
			require.NoError(t, err)

			assert.Equal(t, want.Translated, got.Translated,
				"neutralizing must be a plain substitution, changing nothing else")
		})
	}
}
