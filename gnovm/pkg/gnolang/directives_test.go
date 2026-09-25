package gnolang

import (
	"go/ast"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// isDirectiveText copies the unexported go/ast.isDirective, so it owes a
// fidelity check: go/ast strips directives from CommentGroup.Text(), which
// makes the rule observable. A future Go release changing it fails here, which
// is the intent -- following such a change should be a decision, not drift.
func TestIsDirectiveTextMirrorsGoAst(t *testing.T) {
	t.Parallel()

	goSaysDirective := func(text string) bool {
		cg := &ast.CommentGroup{List: []*ast.Comment{
			{Text: "//" + text}, {Text: "//sentinel"},
		}}
		return cg.Text() == "sentinel\n"
	}

	cases := []string{
		"go:build x", "go:generate ls", "go:noinline", "go:embed f", "line f:1:1",
		"line ", "extern x", "export F", "go:", ":", "a:b", "1:2", "A:b", "a:B",
		" a:b", "a :b", "a: b", "see: docs", "TODO:fix", "nolint:all", "x", "://",
	}
	alphabet := []rune("ab:9 -_/\tAZ.")
	r := rand.New(rand.NewSource(1))
	for range 20000 {
		var sb strings.Builder
		for range r.Intn(8) {
			sb.WriteRune(alphabet[r.Intn(len(alphabet))])
		}
		cases = append(cases, sb.String())
	}
	for _, c := range cases {
		if strings.TrimSpace(c) == "" {
			continue // indistinguishable from a dropped comment via Text()
		}
		assert.Equal(t, goSaysDirective(c), isDirectiveText(c),
			"disagrees with go/ast.isDirective on %q", c)
	}
}

func TestIsDirectiveComment(t *testing.T) {
	t.Parallel()

	for text, want := range map[string]bool{
		"//go:build ignore":    true,
		"//go:generate ls":     true,
		"//go:noinline":        true,
		"//line f.gno:9:1":     true,
		"//export F":           true,
		"/*line f.gno:9:1*/":   true, // Go honours the block form anywhere
		"/*line prose*/":       true, // prefix-only, as go/ast is for "//line"
		"/* ordinary */":       false,
		"/*linefoo*/":          false,
		"// ordinary comment":  false,
		"// see: the docs":     false, // the space makes it prose
		"//TODO:fix":           false, // uppercase is not a tool name
		"//nolint:gosec":       true,  // a directive by Go's rule; see the transpiler
		"not a comment at all": false,
	} {
		assert.Equal(t, want, IsDirectiveComment(text), "%q", text)
	}
}
