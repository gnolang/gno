package gnolang

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// The reserved-name rule belongs to the deploy gate, so every shape it
// refuses must be refused HERE. The companion test in machine_test.go pins
// the other half: preprocess, which re-runs stored packages at node boot,
// must accept them all.
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
		// `:=` with one new variable ASSIGNS cur, so go/types records it in
		// Uses, not Defs; checkDeclNames alone never sees this shape.
		{"cur same-scope rebind", "func F(cur realm) string { cur, y := cur, 1; _ = y; return cur.PkgPath() }", "cannot reassign the crossing"},
		// The shape preprocess must still accept for stored packages: a local
		// `cur` threaded from a crossing function with a blank first param.
		{"cur threaded from blank param", "func G(cur realm) string { return cur.PkgPath() }\nfunc F(_ realm, c realm) string { cur := c; return G(cur) }", "contextual keyword"},

		// The one legal binding, and the shapes the rule must not touch.
		{"crossing first param", "func F(cur realm) {}", ""},
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
			}, TypeCheckOptions{Getter: mockPackageGetter{}, Mode: TCLatestStrict, ProdOnly: true})
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

// Source the chain already admitted must keep type-checking. Node start
// re-checks every stored stdlib from state, and `math/modf.gno` as written
// at genesis names a result `int` — so without this the binary that adds the
// rule cannot start on any chain that predates it, which is a genesis replay.
func TestCheckDeclNamesSkipsSourceAlreadyOnChain(t *testing.T) {
	// The real shape: Go's own Modf signature, which the stdlib used.
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
