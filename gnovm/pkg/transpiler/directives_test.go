package transpiler

import (
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
