package transpiler

import (
	"math/rand"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// TestNoGoGenerateSurvivesTranspile is a differential test of the chain's
// directive rule (gno.FindDirectiveComment, adr/pr6078_forbid_directives.md)
// against the transpiler: no file the rule accepts may come out of Transpile
// with a column-1 go:generate line, which `go generate` would run.
//
// Random rather than enumerated on purpose. Each of the three hand-written
// mirrors this rule was checked against missed a shape go/printer rewrites
// (indentation, control bytes, a doc comment's first line); this samples the
// shapes instead of listing the ones already thought of. The seed is fixed so a
// failure reproduces; the count is sized to run in about a second.
func TestNoGoGenerateSurvivesTranspile(t *testing.T) {
	t.Parallel()

	pieces := []string{
		"//go:generate echo x", "\n", "\n", " ", "\t", "\r", "\v", "\f", "\x01",
		"*", "/", " ", "-", "#", "1.", "[a]:", "ok", "  ", "\t\t",
	}
	frag := func(r *rand.Rand) string {
		var b strings.Builder
		for n := r.Intn(8); n >= 0; n-- {
			b.WriteString(pieces[r.Intn(len(pieces))])
		}
		return b.String()
	}
	// Positions where go/printer treats a comment or literal differently.
	wraps := []func(string) string{
		func(s string) string { return "package zz\n\n/*" + s + "*/\nvar X = 1\n" },                // decl doc
		func(s string) string { return "/*" + s + "*/\npackage zz\n" },                             // package doc
		func(s string) string { return "package zz\n\nvar X = 1\n\n/*" + s + "*/\n" },              // floating
		func(s string) string { return "package zz\n\nfunc F() {\n\t/*" + s + "*/\n\t_ = 0\n}\n" }, // in a body
		func(s string) string { return "package zz\n\ntype T struct {\n\t/*" + s + "*/\n\tA int\n}\n" },
		func(s string) string { return "package zz\n\nconst S = `" + s + "`\n" }, // raw string
	}

	r := rand.New(rand.NewSource(6078))
	accepted := 0
	for i := 0; i < 20000; i++ {
		src := wraps[r.Intn(len(wraps))](frag(r))
		if _, found, err := gno.FindDirectiveComment(src); found || err != nil {
			continue
		}
		res, err := Transpile(src, "gno", "zz.gno")
		if err != nil {
			continue // not a file that transpiles, so nothing for go generate
		}
		accepted++
		for line := range strings.SplitSeq(res.Translated, "\n") {
			if strings.HasPrefix(line, "//go:generate ") || strings.HasPrefix(line, "//go:generate\t") {
				t.Fatalf("accepted %q, but its transpiled output runs under go generate:\n%s", src, res.Translated)
			}
		}
	}
	if accepted < 1000 {
		t.Fatalf("only %d accepted files were transpiled; the generator no longer exercises the rule", accepted)
	}
}
