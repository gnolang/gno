package test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// A file whose declared types all agree carries no block and stays Clean;
// the canonical spelling covers rune/byte, untyped constants, declared,
// pointer, map, function and interface types.
func TestTypeDiff_AgreeingFileHasNoBlock(t *testing.T) {
	o := errorcheckOpts(t)
	src := `// compile

package p

var x = 1
var s string = "a"
const c = 'a'
const d = 1.5
const e, b = "s", true

type T struct {
	a int
	b []byte
}

func f(a int, b ...string) (r int32) {
	y := a
	_ = y
	for i, v := range b {
		_, _ = i, v
	}
	var tp *T = &T{}
	_ = tp
	if z := byte(1); z > 0 {
	}
	var m map[string]func(int) error
	_ = m
	var it interface{ M(int) string }
	_ = it
	fl := func(q rune) {}
	_ = fl
	return 0
}

func (t T) M() {}
`
	out, err := runFiletest(o, "testdata/compile/agree.go", src)
	if err != nil || strings.Contains(out, "GnoTypeDiffers") || !strings.HasSuffix(out, "\n// Verdict: Clean compile\n") {
		t.Fatalf("err=%v out:\n%s", err, out)
	}
}

// `'a' + 1<<31 - 1<<31` is an untyped rune in Go; Gno folds a rune and
// an int beyond int32 as int (#6389): one block line, the umbrella
// KnownIssue, and a hand-written per-issue ref is reconciled like an
// errorcheck gap group.
func TestTypeDiff_RuneArithmetic(t *testing.T) {
	o := errorcheckOpts(t)
	src := "// compile\n\npackage p\n\nvar x = 'a' + 1<<31 - 1<<31\nconst k = 'a' + 1<<31 - 1<<31\n"
	out, err := runFiletest(o, "testdata/compile/rune.go", src)
	if err != nil {
		t.Fatal(err)
	}
	want := src + "\n// GnoTypeDiffers:\n// line 5: x gno int, go int32\n// line 6: k gno untyped int, go untyped rune\n\n// Verdict: KnownIssue compile: #6344 types differ: lines 5, 6\n"
	if out != want {
		t.Fatalf("sync wrote:\n%s\nwant:\n%s", out, want)
	}
	o.Sync = false
	if _, err := runFiletest(o, "testdata/compile/rune.go", out); err != nil {
		t.Fatalf("verify of the synced file: %v", err)
	}
	// A per-issue ref written by hand owns its lines; an unlisted line is
	// stale for verify and goes to the umbrella under sync.
	hand := strings.Replace(out, "#6344 types differ: lines 5, 6", "#6389 rune arithmetic; types differ: lines 5", 1)
	_, err = runFiletest(o, "testdata/compile/rune.go", hand)
	if err == nil || !strings.Contains(err.Error(), "hand-written verdict is stale: lines 6 are types differ but unlisted") {
		t.Fatalf("stale hand-written list accepted: %v", err)
	}
	o.Sync = true
	out, err = runFiletest(o, "testdata/compile/rune.go", hand)
	if err != nil || !strings.HasSuffix(out, "// Verdict: KnownIssue compile: #6389 rune arithmetic; types differ: lines 5; #6344 types differ: lines 6\n") {
		t.Fatalf("sync of the hand-written verdict: err=%v out:\n%s", err, out)
	}
	o.Sync = false
	if _, err := runFiletest(o, "testdata/compile/rune.go", out); err != nil {
		t.Fatalf("verify of the reconciled file: %v", err)
	}
}

// A non-constant shift's untyped operand takes the type of its context
// (#6393) and `'a' + 1` is a rune (#6339): both sides agree, no block.
func TestTypeDiff_ShiftContext(t *testing.T) {
	o := errorcheckOpts(t)
	src := "// compile\n\npackage p\n\nvar s uint\nvar r = 'a' + 1\n\nfunc f() {\n\tx := 1 + ('a' << s)\n\tvar y uint8 = 1 + ('a' << s)\n\t_, _ = x, y\n}\n"
	out, err := runFiletest(o, "testdata/compile/shift.go", src)
	if err != nil || strings.Contains(out, "GnoTypeDiffers") || !strings.HasSuffix(out, "\n// Verdict: Clean compile\n") {
		t.Fatalf("err=%v out:\n%s", err, out)
	}
}

// A block the types no longer justify fails verify in every mode, as a
// Clean verdict over a mismatch does.
func TestTypeDiff_StaleBlockFailsVerify(t *testing.T) {
	o := errorcheckOpts(t)
	o.Sync = false
	for _, tc := range []struct{ name, src, want string }{
		{"compile", "// compile\n\npackage p\n\nvar x = 1\n\n// GnoTypeDiffers:\n// line 5: x gno int, go int32\n\n// Verdict: KnownIssue compile: #6344 types differ: lines 5\n", "stale `// GnoTypeDiffers:` block"},
		{"run", "package main\n\nvar x = 1\n\nfunc main() { println(x) }\n\n// GnoOutput:\n// 1\n\n// GoOutput:\n// 1\n\n// GnoTypeDiffers:\n// line 3: x gno int, go int32\n\n// Verdict: KnownIssue run: #6344 types differ: lines 3\n", "stale `// GnoTypeDiffers:` block"},
		{"run clean over a mismatch", "package main\n\nvar x = 'a' + 1<<31 - 1<<31\n\nfunc main() { println(x) }\n\n// GnoOutput:\n// 97\n\n// GoOutput:\n// 97\n\n// Verdict: Clean run\n", "missing `// GnoTypeDiffers:` golden"},
		{"errorcheck", "package p\n\nvar x = 1\nvar y int = \"s\" // ERROR \"cannot\"\n\n// GnoError:\n// line 4: cannot use untyped string as IntKind\n\n// GnoTypeDiffers:\n// line 3: x gno int, go int32\n\n// Verdict: KnownIssue errorcheck: #6344 types differ: lines 3\n", "stale `// GnoTypeDiffers:` block"},
	} {
		_, err := runFiletest(o, "testdata/"+tc.name+"/stale.go", tc.src)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: want %q, got %v", tc.name, tc.want, err)
		}
	}
}

// Run mode: a file whose outputs match but whose declared types differ
// is a KnownIssue, written by sync and accepted by verify.
func TestTypeDiff_RunMode(t *testing.T) {
	o := errorcheckOpts(t)
	src := "package main\n\nvar x = 'a' + 1<<31 - 1<<31\n\nfunc main() {\n\ty := 'b' + 1<<31 - 1<<31\n\tprintln(x, y)\n}\n"
	out, err := runFiletest(o, "testdata/run/r.go", src)
	if err != nil {
		t.Fatal(err)
	}
	want := src + "\n// GnoOutput:\n// 97 98\n\n// GoOutput:\n// 97 98\n\n// GnoTypeDiffers:\n// line 3: x gno int, go int32\n// line 6: y gno int, go int32\n\n// Verdict: KnownIssue run: #6344 types differ: lines 3, 6\n"
	if out != want {
		t.Fatalf("sync wrote:\n%s\nwant:\n%s", out, want)
	}
	o.Sync = false
	if _, err := runFiletest(o, "testdata/run/r.go", out); err != nil {
		t.Fatalf("verify of the synced file: %v", err)
	}
}

// The canonical spelling both sides are normalised to.
func TestTypeStrings(t *testing.T) {
	const pkg = "gno.land/p/x"
	errT := &gno.DeclaredType{PkgPath: ".uverse", Name: "error"}
	strT := &gno.DeclaredType{PkgPath: "strings", Name: "Builder"}
	for want, typ := range map[string]gno.Type{
		"int32":             gno.Int32Type,
		"uint8":             gno.Uint8Type,
		"untyped int":       gno.UntypedBigintType,
		"untyped rune":      gno.UntypedRuneType,
		"untyped float":     gno.UntypedBigdecType,
		"untyped string":    gno.UntypedStringType,
		"untyped bool":      gno.UntypedBoolType,
		"T":                 &gno.DeclaredType{PkgPath: pkg, Name: "T"},
		"strings.Builder":   strT,
		"error":             errT,
		"*[]uint8":          &gno.PointerType{Elt: &gno.SliceType{Elt: gno.Uint8Type}},
		"map[string][3]int": &gno.MapType{Key: gno.StringType, Value: &gno.ArrayType{Len: 3, Elt: gno.IntType}},
		"func(int, ...string) (int32, error)": &gno.FuncType{
			Params:  []gno.FieldType{{Name: "a", Type: gno.IntType}, {Name: "b", Type: &gno.SliceType{Elt: gno.StringType, Vrd: true}}},
			Results: []gno.FieldType{{Type: gno.Int32Type}, {Type: errT}},
		},
		"struct{a int; b []uint8}": &gno.StructType{Fields: []gno.FieldType{{Name: "a", Type: gno.IntType}, {Name: "b", Type: &gno.SliceType{Elt: gno.Uint8Type}}}},
		"interface{M(int) string; N()}": &gno.InterfaceType{Methods: []gno.FieldType{
			{Name: "N", Type: &gno.FuncType{}},
			{Name: "M", Type: &gno.FuncType{Params: []gno.FieldType{{Type: gno.IntType}}, Results: []gno.FieldType{{Type: gno.StringType}}}},
		}},
	} {
		if got := typeString(typ, pkg); got != want {
			t.Errorf("gno %v: got %q, want %q", typ, got, want)
		}
	}
	src := `package x
import "strings"
var _ strings.Builder
type T struct{}
var (
	a rune
	b byte
	c = 'a'
	d = 1
	e = 1.5
	f = "s"
	g = true
	h T
	i error
	j *[]byte
	k map[string][3]int
	l func(p1 int, p2 ...string) (int32, error)
	m struct{ fa int; fb []byte }
	n interface{ N(); M(int) string }
	o strings.Builder
	p any
)
const q = 'a' + 1
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}}
	cfg := types.Config{Importer: stringsImporter{}}
	if _, err := cfg.Check(pkg, fset, []*ast.File{file}, info); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"a": "int32", "b": "uint8", "c": "int32", "d": "int", "e": "float64", "f": "string", "g": "bool",
		"h": "T", "i": "error", "j": "*[]uint8", "k": "map[string][3]int",
		"l": "func(int, ...string) (int32, error)", "m": "struct{fa int; fb []uint8}",
		"n": "interface{M(int) string; N()}", "o": "strings.Builder", "p": "interface{}", "q": "untyped rune",
	}
	for id, obj := range info.Defs {
		w, ok := want[id.Name]
		if !ok || obj == nil {
			continue
		}
		if got := goTypeString(obj.Type(), pkg); got != w {
			t.Errorf("go %s: got %q, want %q", id.Name, got, w)
		}
	}
}

// stringsImporter resolves the one import TestTypeStrings needs from a
// stub, so the test shells out to nothing.
type stringsImporter struct{}

func (stringsImporter) Import(path string) (*types.Package, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path+".go", "package strings\ntype Builder struct{}\n", 0)
	if err != nil {
		return nil, err
	}
	return (&types.Config{}).Check(path, fset, []*ast.File{file}, nil)
}
