package strings

import (
	"testing"
)

func (r *Replacer) PrintTrie() string {
	r.buildOnce()
	gen := r.r.(*genericReplacer)
	return gen.printNode(&gen.root, 0)
}

func (r *genericReplacer) printNode(t *trieNode, depth int) (s string) {
	if t.priority > 0 {
		s += "+"
	} else {
		s += "-"
	}
	s += "\n"

	if t.prefix != "" {
		s += Repeat(".", depth) + t.prefix
		s += r.printNode(t.next, depth+len(t.prefix))
	} else if t.table != nil {
		for b, m := range r.mapping {
			if int(m) != r.tableSize && t.table[m] != nil {
				s += Repeat(".", depth) + string([]byte{byte(b)})
				s += r.printNode(t.table[m], depth+1)
			}
		}
	}
	return
}

func TestGenericTrieBuilding(t *testing.T) {
	testCases := []struct{ in, out string }{
		{"abc;abdef;abdefgh;xx;xy;z", `-
			a-
			.b-
			..c+
			..d-
			...ef+
			.....gh+
			x-
			.x+
			.y+
			z+
			`},
		{"abracadabra;abracadabrakazam;abraham;abrasion", `-
			a-
			.bra-
			....c-
			.....adabra+
			...........kazam+
			....h-
			.....am+
			....s-
			.....ion+
			`},
		{"aaa;aa;a;i;longerst;longer;long;xx;x;X;Y", `-
			X+
			Y+
			a+
			.a+
			..a+
			i+
			l-
			.ong+
			....er+
			......st+
			x+
			.x+
			`},
		{"foo;;foo;foo1", `+
			f-
			.oo+
			...1+
			`},
	}

	for _, tc := range testCases {
		keys := Split(tc.in, ";")
		args := make([]string, len(keys)*2)
		for i, key := range keys {
			args[i*2] = key
		}

		got := NewReplacer(args...).PrintTrie()
		// Remove tabs from tc.out
		wantbuf := make([]byte, 0, len(tc.out))
		for i := 0; i < len(tc.out); i++ {
			if tc.out[i] != '\t' {
				wantbuf = append(wantbuf, tc.out[i])
			}
		}
		want := string(wantbuf)

		if got != want {
			t.Errorf("PrintTrie(%q)\ngot\n%swant\n%s", tc.in, got, want)
		}
	}
}
