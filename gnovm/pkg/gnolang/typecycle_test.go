package gnolang

import "testing"

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
