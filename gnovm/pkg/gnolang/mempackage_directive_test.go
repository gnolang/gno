package gnolang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Directives carry no meaning in Gno, so a submitted package may not declare
// one: an honoring consumer turns the tag into behaviour the submitter
// controls (see adr/pr5978_typecheck_strip_file_goversion.md), and an inert tag
// misleads whoever audits the stored source.
func TestFindDirectiveComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want bool
	}{
		{"no constraint", "package zz\n\nfunc F() {}\n", false},
		{"go:build", "//go:build go1.22\n\npackage zz\n", true},
		{"go:build ignore", "//go:build ignore\n\npackage zz\n", true},
		{"legacy +build", "// +build linux\n\npackage zz\n", true},
		{"both forms", "//go:build linux\n// +build linux\n\npackage zz\n", true},
		{"after doc comment", "// Package z does things.\n//go:build ignore\n\npackage zz\n", true},
		{"indented", "   //go:build ignore\n\npackage zz\n", true},
		{"crlf line endings", "//go:build ignore\r\n\r\npackage zz\r\n", true},

		// False-positive guards: only the header is a constraint position, so
		// a package that merely mentions the syntax stays valid. A linter or
		// formatter realm written in Gno is exactly such a package.
		{"inside string literal", "package zz\n\nconst s = \"//go:build ignore\"\n", false},
		{"ordinary comment with colon", "package zz\n\n// see: the docs\nfunc F() {}\n", false},
		{"uppercase pseudo-directive", "package zz\n\n//TODO:fix\nfunc F() {}\n", false},
		{"not a constraint comment", "// go:build ignore\n\npackage zz\n", false},
		{"go:embed directive", "//go:embed x.txt\n\npackage zz\n", true},

		// Directives outside the header: pragmas attach to declarations, and a
		// line directive may sit anywhere. All are meaningless to the VM, and
		// //line additionally forges the position reported in a failed tx.
		{"pragma mid-file", "package zz\n\n//go:noinline\nfunc F() {}\n", true},
		{"line directive", "package zz\n\n//line /etc/passwd:99:1\nfunc F() {}\n", true},
		{"cgo export", "package zz\n\n//export F\nfunc F() {}\n", true},
		{"trailing directive", "package zz\n\nvar x = 1 //go:generate rm -rf /\n", true},

		// //nolint is the one allowed directive: golangci-lint never reads
		// .gno, so it is as meaningless as the rest, but no consumer of the
		// stored source honours it either, and Go authors arrive with the
		// habit. See isAllowedDirective.
		{"nolint bare", "package zz\n\n//nolint\nfunc F() {}\n", false},
		{"nolint with linters", "package zz\n\n//nolint:gosec,errcheck\nfunc F() {}\n", false},
		{"nolint with reason", "package zz\n\n//nolint:gosec // why\nfunc F() {}\n", false},
		{"nolint trailing", "package zz\n\nvar x = 1 //nolint:all\n", false},
		{"nolint lookalike is not allowed", "package zz\n\n//nolintfoo:bar\nfunc F() {}\n", true},

		// Go honours the BLOCK form of a line directive anywhere in a file,
		// not only at the start of a line (go/scanner accepts a comment when
		// lit[1] == '*' || offs == lineOffset). A "//"-only guard misses it.
		{"block line directive", "package zz\n\n/*line forged.gno:999:1*/\nfunc F() {}\n", true},
		{"block line mid-line", "package zz\n\nvar x = 1 /*line f.gno:9:1*/\n", true},
		{"block line in header", "/*line f.gno:9:1*/\npackage zz\n", true},
		{"ordinary block comment", "package zz\n\n/* hello */\nfunc F() {}\n", false},
		{"block without trailing space", "package zz\n\n/*linefoo*/\nfunc F() {}\n", false},

		// Bypasses of a naive line scan. go/parser honours the constraint in
		// every case below (verified against go1.25.9), so missing one would
		// let a submitter keep a live tag in a stored file.
		{"BOM before constraint", "\ufeff//go:build ignore\n\npackage zz\n", true},
		{"package inside block comment", "/*\npackage zz\n*/\n//go:build ignore\n\npackage zz\n", true},
		{"block comment first", "/* hi */\n//go:build ignore\n\npackage zz\n", true},
		{"no blank line before clause", "//go:build ignore\npackage zz\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, got := FindDirectiveComment(tt.body)
			assert.Equal(t, tt.want, got)
		})
	}
}

// The security property, stated executably: if go/parser honours a build
// constraint in a body, FindDirectiveComment must find it. A gap would leave a
// live tag in stored source, which is the whole failure this rule prevents.
//
// Built by combining the fragments that make the predicate hard — a BOM, block
// comments, a "package" line inside one, empty comments — because the bypasses
// found while writing this rule all came from that shape rather than from an
// exotic directive spelling.
func TestNoHonouredConstraintEscapes(t *testing.T) {
	t.Parallel()

	frag := []string{
		"//go:build go1.9", "// +build x", "//line a:1:1", "//go:noinline",
		"\ufeff", "/*", "*/", "package zz", "func F(){}", "\n", " ", "//", "/**/", "\t", "//x",
		"/*line f.gno:9:1*/", "/*line ", "var x = 1",
	}
	r := rand.New(rand.NewSource(7))
	for range 100000 {
		var sb strings.Builder
		for range r.Intn(7) {
			sb.WriteString(frag[r.Intn(len(frag))])
			sb.WriteString("\n")
		}
		body := sb.String()
		fset := token.NewFileSet()
		gof, err := parser.ParseFile(fset, "x.gno", body,
			parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			continue // unparseable files are rejected before the directive check
		}
		_, found := FindDirectiveComment(body)
		if gof.GoVersion != "" && !found {
			t.Fatalf("go/parser honours GoVersion=%q but the predicate missed it in %q",
				gof.GoVersion, body)
		}
		// Same property for line directives, observed through their effect:
		// if any position in the file resolves to a filename other than the
		// one handed to the parser, a line directive was honoured. This is
		// what catches the block form, which carries no GoVersion.
		if !found {
			for _, d := range gof.Decls {
				if got := fset.Position(d.Pos()).Filename; got != "x.gno" {
					t.Fatalf("a line directive remapped positions to %q but the predicate missed it in %q",
						got, body)
				}
			}
		}
	}
}

// isDirectiveText is a copy of the unexported go/ast.isDirective, so it owes a
// fidelity check: too narrow and a directive Go honours slips through into
// stored source, too broad and legitimate packages stop deploying.
//
// go/ast.isDirective is unexported but observable — CommentGroup.Text() drops
// directives — so compare against that. If a future Go release changes the
// rule, this test fails, which is the intent: the copy is pinned deliberately
// (a consensus rule must not shift under a toolchain upgrade), and following a
// change should be a decision rather than a silent drift.
func TestIsDirectiveTextMirrorsGo(t *testing.T) {
	t.Parallel()

	goSaysDirective := func(text string) bool {
		cg := &ast.CommentGroup{List: []*ast.Comment{
			{Text: "//" + text}, {Text: "//sentinel"},
		}}
		return cg.Text() == "sentinel\n" // the candidate line was dropped
	}

	cases := []string{
		"go:build x", "go:generate ls", "go:noinline", "go:embed f", "line f:1:1",
		"line ", "extern x", "export F", "go:", ":", "a:b", "1:2", "A:b", "a:B",
		" a:b", "a :b", "a: b", "see: docs", "TODO:fix", "nolint:all", "go:build",
		"lineX f", "exportF", "extern", "export", "line", "go1:x", "go_:x", "ab0:9",
		"x", "://", "a::b",
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
		// An empty or whitespace-only comment is indistinguishable from a
		// dropped one via Text(), so it cannot be compared this way.
		if strings.TrimSpace(c) == "" {
			continue
		}
		assert.Equal(t, goSaysDirective(c), isDirectiveText(c),
			"disagrees with go/ast.isDirective on %q", c)
	}
}

// The scan walks to EOF on every .gno file of every submitted package, so a
// scanner that failed to advance would be an unbounded loop on the AddPackage
// path rather than a cosmetic bug. Malformed input reaches it: validation runs
// on attacker-supplied bytes.
func TestFindDirectiveComment_Terminates(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"unterminated block comment": "/* never closed\n//go:noinline\npackage zz\n",
		"unterminated string":        "package zz\nvar s = \"unclosed\n",
		"illegal bytes":              "package zz\n\x00\x01\xff\xfe\n",
		"lone slash":                 "/",
		"lone slash star":            "/*",
		"empty":                      "",
		"invalid utf8 in comment":    "package zz\n// \xff\xfe\xfd\n",
		"many comments":              strings.Repeat("//c\n", 50000) + "package zz\n",
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			done := make(chan struct{})
			go func() { defer close(done); FindDirectiveComment(body) }()
			select {
			case <-done:
			case <-time.After(30 * time.Second):
				t.Fatal("FindDirectiveComment did not terminate")
			}
		})
	}
}

func validateBody(t *testing.T, mptype MemPackageType, pkgPath, body string) error {
	t.Helper()
	name := pkgPath[strings.LastIndex(pkgPath, "/")+1:]
	return ValidateMemPackageAny(&std.MemPackage{
		Type:  mptype,
		Name:  name,
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "z.gno", Body: body}},
	})
}

func TestValidateMemPackage_Directives(t *testing.T) {
	t.Parallel()

	const userPath = "gno.land/p/demo/zz"

	t.Run("user package with a constraint is rejected", func(t *testing.T) {
		t.Parallel()

		// Guard: the same package without the tag must validate, so a failure
		// below cannot be misread as the package being invalid for some other
		// reason.
		require.NoError(t, validateBody(t, MPUserProd, userPath, "package zz\nfunc F() {}\n"))

		for _, body := range []string{
			"//go:build go1.22\n\npackage zz\nfunc F() {}\n",
			"//go:build ignore\n\npackage zz\nfunc F() {}\n",
			"// +build linux\n\npackage zz\nfunc F() {}\n",
		} {
			err := validateBody(t, MPUserProd, userPath, body)
			assert.ErrorContains(t, err, "directives are not supported",
				"a submitted package must not carry a build constraint: %q", body)
		}
	})

	t.Run("every user type is covered", func(t *testing.T) {
		t.Parallel()

		// Test files are stored on chain too (in the #allbutprod sibling) and
		// are just as submitter-controlled as production files.
		const body = "//go:build ignore\n\npackage zz\nfunc F() {}\n"
		for _, mptype := range []MemPackageType{MPUserAll, MPUserProd, MPUserTest} {
			err := validateBody(t, mptype, userPath, body)
			assert.ErrorContains(t, err, "directives are not supported",
				"type %v must reject build constraints", mptype)
		}

		// MPUserIntegration is the fourth IsUserlib type; its path must carry
		// the "_test" suffix (see MemPackageType.Validate).
		err := validateBody(t, MPUserIntegration, userPath+"_test",
			"//go:build ignore\n\npackage zz_test\nfunc F() {}\n")
		assert.ErrorContains(t, err, "directives are not supported",
			"type %v must reject build constraints", MPUserIntegration)
	})

	t.Run("a tagged file anywhere in the package is found", func(t *testing.T) {
		t.Parallel()

		// Every file is checked, not just the first, and the error names the
		// offending one so the author knows which to fix.
		err := ValidateMemPackageAny(&std.MemPackage{
			Type: MPUserProd,
			Name: "zz",
			Path: userPath,
			Files: []*std.MemFile{
				{Name: "a.gno", Body: "package zz\nfunc A() {}\n"},
				{Name: "b.gno", Body: "//go:build ignore\n\npackage zz\nfunc B() {}\n"},
			},
		})
		assert.ErrorContains(t, err, "directives are not supported")
		assert.ErrorContains(t, err, `"b.gno"`, "the error must name the tagged file")
		assert.NotContains(t, fmt.Sprint(err), "a.gno", "the clean file must not be blamed")
	})

	t.Run("nolint is allowed", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, validateBody(t, MPUserProd, userPath,
			"package zz\n\n//nolint:gosec\nfunc F() {}\n"))
	})

	t.Run("a directive cannot rewrite its own error", func(t *testing.T) {
		t.Parallel()

		// The check runs before the file is parsed: a line directive rewrites
		// the positions go/parser reports, so parsing first would let a
		// rejected file choose the filename and line printed in its own error.
		err := validateBody(t, MPUserProd, userPath, "//line /etc/passwd:999:1\npackag zz\n")
		assert.ErrorContains(t, err, "directives are not supported")
		assert.NotContains(t, fmt.Sprint(err), "/etc/passwd",
			"the error must not carry the forged path")
	})

	t.Run("control bytes in a directive are escaped", func(t *testing.T) {
		t.Parallel()

		// The directive name is submitted text, and the error reaches a
		// terminal (gno lint) and a transaction result.
		err := validateBody(t, MPUserProd, userPath,
			"package zz\n\n//a:b\x1b[2J\x07evil\nfunc F() {}\n")
		require.Error(t, err)
		assert.NotContains(t, fmt.Sprint(err), "\x1b", "escape byte must not be echoed raw")
		assert.NotContains(t, fmt.Sprint(err), "\x07", "bell byte must not be echoed raw")
	})

	t.Run("stdlibs are not affected", func(t *testing.T) {
		t.Parallel()

		// Stdlibs ship with the node binary rather than being submitted, and
		// the VM filetest suite pins that constraints are inert
		// (gnovm/tests/files/build0.gno, extern/ct).
		err := validateBody(t, MPStdlibProd, "strings", "//go:build ignore\n\npackage strings\nfunc F() {}\n")
		assert.NoError(t, err)
	})
}
