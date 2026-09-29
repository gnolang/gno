package gnolang

import (
	"fmt"
	"strings"
	"testing"
)

func TestTypeDeclInvalidCycle(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"direct self", "type S struct{ s S }", "S -> S"},
		{"pointer breaks self", "type L struct{ next *L }", ""},
		{"pointer and defined type", "type ( P = *T; T P )", ""},
		{"alias only through slice", "type ( A = []B; B = A )", "A -> B -> A"},
		{"alias only through func", "type A = func() A", "A -> A"},
		{"embedded interface", "type I interface{ I }", "I -> I"},
		{"method signature", "type I interface{ M() I }", ""},
		{"mixed alias group", "type ( e = f; f = g; g = []h; h i; i = j; j = e )", ""},
		{"array closes alias group", "type ( e = f; f = g; g = [3]h; h i; i = j; j = e )", "e -> f -> g -> h -> i -> j -> e"},
		// The first edge into B is indirect, so a DFS that judged the
		// cycle it happened to close would call the group legal.
		{"direct cycle beside a legal one", "type A struct{ b *B; c C }\ntype C struct{ b B }\ntype B struct{ a A }", "A -> C -> B -> A"},
		{"other package is a leaf", "type A struct{ t std.Address }", ""},
	}
	m := NewMachine("main", nil)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fn := m.MustParseFile("main.gno", "package main\n"+c.src+"\n")
			g := newTypeDeclGraph(appendTypeDeclSites(nil, fn, fn.Decls))
			got := Names(g.invalidCycle()).Join(" -> ")
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

// preprocessSrc runs src as a package through predefinition and
// preprocessing, returning the panic value if any.
func preprocessSrc(src string) (failure any) {
	defer func() { failure = recover() }()
	m := NewMachine("main", nil)
	n := m.MustParseFile("main.gno", "package main\n\n"+src+"\n\nfunc main() {}\n")
	m.RunFiles(n)
	return nil
}

// TestTypeDeclCyclesPredefine runs cyclic declaration groups through
// predefinition. A bare-name reference into a cycle, `type A B` with B
// pointing back through an indirection, used to reach a member before it
// had a slot, whichever declaration came first.
func TestTypeDeclCyclesPredefine(t *testing.T) {
	type tc struct {
		name, src string
		invalid   bool
	}
	var cases []tc
	shapes := []struct{ name, tmpl string }{{"ptr", "*A"}, {"slice", "[]A"}, {"func", "func() A"}}
	eq := func(alias bool) string {
		if alias {
			return "= "
		}
		return ""
	}
	for _, sh := range shapes {
		for _, aAlias := range []bool{false, true} {
			for _, bAlias := range []bool{false, true} {
				declA := "type A " + eq(aAlias) + "B"
				declB := "type B " + eq(bAlias) + sh.tmpl
				tag := sh.name + "-" + eq(aAlias) + "A-" + eq(bAlias) + "B"
				invalid := aAlias && bAlias // alias-only cycle
				cases = append(cases,
					tc{tag + "-AB", declA + "\n" + declB, invalid},
					tc{tag + "-BA", declB + "\n" + declA, invalid})
			}
		}
	}
	cases = append(cases,
		tc{"alias chain to slice", "type (\n\ta = b\n\tb = c\n\tc []a\n)", false},
		tc{"alias chain to slice, reversed", "type (\n\tc []a\n\ta = b\n\tb = c\n)", false},
		tc{"defined over defined, pointer back", "type B struct{ p *A }\ntype A B", false},
		tc{"alias over defined, pointer back", "type A = B\ntype B struct{ p *A }", false},
		tc{"alias over defined, pointer back, reversed", "type B struct{ p *A }\ntype A = B", false},
		tc{"defined chain of three", "type C struct{ p *A }\ntype B C\ntype A B", false},
		tc{"defined over defined, direct back", "type B struct{ a A }\ntype A B", true},
	)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			failure := preprocessSrc(c.src)
			switch {
			case c.invalid && failure == nil:
				t.Fatalf("accepted an invalid cycle:\n%s", c.src)
			case c.invalid && !strings.Contains(fmt.Sprint(failure), "invalid recursive type"):
				t.Fatalf("wrong error %v for:\n%s", failure, c.src)
			case !c.invalid && failure != nil:
				t.Fatalf("rejected a legal cycle with %v:\n%s", failure, c.src)
			}
		})
	}
}
