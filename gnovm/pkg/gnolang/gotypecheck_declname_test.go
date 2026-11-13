package gnolang

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// The reserved-name rule belongs to the deploy gate, so every shape it
// refuses must be refused HERE. The companion tests in
// reserved_names_boot_test.go pin the other half: preprocess, which re-runs
// stored packages at node boot, must accept them all.
func TestCheckDeclNames(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"param", "func F(len int) int { return len }", "cannot be shadowed: len"},
		{"result", "func F(f float64) (int float64, frac float64) { return 1, 0.5 }", "cannot be shadowed: int"},
		{"uverse param", "func F(address string) string { return address }", "cannot be shadowed: address"},
		{"receiver", "type T struct{}\nfunc (string T) M() {}", "cannot be shadowed: string"},
		{"range", "func F(xs []int) { for len := range xs { _ = len } }", "cannot be shadowed: len"},
		{"type switch", "func F(a any) { switch len := a.(type) { default: _ = len } }", "cannot be shadowed: len"},
		{"local", "func F() { len := 3; _ = len }", "cannot be shadowed: len"},
		{"package level", "var len = 3", "cannot be shadowed: len"},
		{"func type param", "type S struct{ F func(len int) int }", "cannot be shadowed: len"},
		{"interface method", "type I interface{ M(len int) int }", "cannot be shadowed: len"},
		{"closure param", "func F() { g := func(panic string) {}; _ = g }", "cannot be shadowed: panic"},
		{"local type", "func F() { type len struct{}; _ = len{} }", "cannot be shadowed: len"},
		{"cur local", "func F() { cur := 1; _ = cur }", "contextual keyword"},
		{"cur non-realm param", "func F(cur int) int { return cur }", "contextual keyword"},
		{"cur func", "func cur() {}", "contextual keyword"},
		{"cur second param", "func F(a int, cur realm) {}", "contextual keyword"},
		// An import binds its package name in file scope; without an alias
		// the binding is implicit, and go/types records it in Implicits.
		{"cur implicit import", "import \"gno.land/p/demo/cur\"\n\nfunc F() bool { return cur.Previous().IsUser() }", "contextual keyword"},
		{"builtin implicit import", "import \"gno.land/p/demo/len\"\n\nfunc F() int { return len.N }", "cannot be shadowed: len"},
		// prepareGoGno0p9 strips main's and init's `cur realm` before go/types
		// runs. Only that literal shape is stripped: any other parameter
		// stays, so the checker sees it and this rule judges it.
		{"main param builtin", "func main(panic func(string)) string { panic(\"x\"); return \"\" }", "cannot be shadowed: panic"},
		{"main param cur int", "func main(cur int) {}", "contextual keyword"},
		// init with any other parameter is a Go signature error first.
		{"init param builtin", "func init(len int) {}", "init must have no arguments"},
		// `:=` with one new variable ASSIGNS cur, so go/types records it in
		// Uses, not Defs; checkDeclNames alone never sees this shape.
		{"cur same-scope rebind", "func F(cur realm) string { cur, y := cur, 1; _ = y; return cur.PkgPath() }", "cannot reassign the crossing"},
		// The shape preprocess must still accept for stored packages: a local
		// `cur` threaded from a crossing function with a blank first param.
		{"cur threaded from blank param", "func G(cur realm) string { return cur.PkgPath() }\nfunc F(_ realm, c realm) string { cur := c; return G(cur) }", "contextual keyword"},

		// The one legal binding, and the shapes the rule must not touch.
		{"crossing first param", "func F(cur realm) {}", ""},
		{"main cur realm", "func main(cur realm) {}", ""},
		{"init cur realm", "func init(cur realm) {}", ""},
		{"aliased import", "import l \"gno.land/p/demo/len\"\n\nfunc F() int { return l.N }", ""},
		{"blank", "func F() { _ = 1 }", ""},
		{"ordinary names", "func F(n int) int { return n }", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := TypeCheckMemPackage(&std.MemPackage{
				Type: MPUserProd, Name: "probe", Path: "gno.land/r/demo/probe",
				Files: []*std.MemFile{
					{Name: "gnomod.toml", Body: "module = \"gno.land/r/demo/probe\"\ngno = \"0.9\"\n"},
					{Name: "a.gno", Body: "package probe\n\n" + tc.body + "\n"},
				},
			}, TypeCheckOptions{Getter: declNameImports, Mode: TCLatestStrict, ProdOnly: true})
			switch {
			case tc.want == "" && err != nil:
				t.Fatalf("refused a legal binding: %v", err)
			case tc.want != "" && err == nil:
				t.Fatal("accepted")
			case tc.want != "" && !strings.Contains(err.Error(), tc.want):
				t.Fatalf("want %q, got %v", tc.want, err)
			}
		})
	}
}

// Packages whose NAME is reserved, for the import cases above. Each is a
// legal package on its own; it is the importer's implicit binding that the
// gate refuses.
var declNameImports = mockPackageGetter{
	reservedNamePkg("cur", "type Fake struct{}\n\nfunc (Fake) IsUser() bool { return true }\n\nfunc Previous() Fake { return Fake{} }"),
	reservedNamePkg("len", "var N = 5"),
}

func reservedNamePkg(name, body string) *std.MemPackage {
	path := "gno.land/p/demo/" + name
	return &std.MemPackage{
		Type: MPUserProd, Name: name, Path: path,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: "module = \"" + path + "\"\ngno = \"0.9\"\n"},
			{Name: name + ".gno", Body: "package " + name + "\n\n" + body + "\n"},
		},
	}
}

// Defs and Implicits are maps, so the offending binding is picked out of a
// randomized range. Two nodes must word the refusal identically.
func TestCheckDeclNamesIsDeterministic(t *testing.T) {
	mpkg := &std.MemPackage{
		Type: MPUserProd, Name: "probe", Path: "gno.land/r/demo/probe",
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: "module = \"gno.land/r/demo/probe\"\ngno = \"0.9\"\n"},
			{Name: "a.gno", Body: `package probe

func A(len int) int { return len }
func B(cap int) int { return cap }
func C(a any) { switch copy := a.(type) { default: _ = copy } }
func D() { cur := 1; _ = cur }
`},
		},
	}
	seen := map[string]int{}
	for range 200 {
		_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{
			Getter: mockPackageGetter{}, Mode: TCLatestStrict, ProdOnly: true,
		})
		if err == nil {
			t.Fatal("accepted")
		}
		seen[err.Error()]++
	}
	if len(seen) != 1 {
		t.Fatalf("refusal worded %d different ways across 200 runs (%v); the message must not depend on map iteration order", len(seen), seen)
	}
}

// Source the chain already admitted must keep type-checking. A hardfork
// replays its history through InitChain, and gnoland-1's history holds
// deploys with builtin-named parameters — so without this the binary that
// adds the rule cannot replay the chain it is upgrading.
func TestCheckDeclNamesSkipsSourceAlreadyOnChain(t *testing.T) {
	// The shape live packages carry: Go's own Modf signature.
	const body = "func Modf(f float64) (int float64, frac float64) { return 1, 0.5 }"
	mpkg := func() *std.MemPackage {
		return &std.MemPackage{
			Type: MPUserProd, Name: "probe", Path: "gno.land/r/demo/probe",
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: "module = \"gno.land/r/demo/probe\"\ngno = \"0.9\"\n"},
				{Name: "a.gno", Body: "package probe\n\n" + body + "\n"},
			},
		}
	}
	base := TypeCheckOptions{Getter: mockPackageGetter{}, Mode: TCLatestStrict, ProdOnly: true}

	if _, err := TypeCheckMemPackage(mpkg(), base); err == nil {
		t.Fatal("a new deploy must still be refused")
	}

	onChain := base
	onChain.AlreadyOnChain = true
	if _, err := TypeCheckMemPackage(mpkg(), onChain); err != nil {
		t.Fatalf("source already on chain must still check: %v", err)
	}
}

// A stdlib is never admitted: LoadStdlib type-checks it straight from the
// binary, and node start re-checks the stored copy. The gate does not run on
// it, so a stdlib keeps Go's own names, and the stored stdlib source, which
// is part of the app hash, never has to change for this rule.
func TestCheckDeclNamesSkipsStdlibs(t *testing.T) {
	mpkg := &std.MemPackage{
		Type: MPStdlibProd, Name: "math", Path: "math",
		Files: []*std.MemFile{
			{Name: "modf.gno", Body: "package math\n\nfunc Modf(f float64) (int float64, frac float64) { return 1, 0.5 }\n"},
		},
	}
	_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{Getter: mockPackageGetter{}, Mode: TCLatestStrict, ProdOnly: true})
	if err != nil {
		t.Fatalf("a stdlib is not an admission: %v", err)
	}
}
