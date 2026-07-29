// PKGPATH: gno.land/r/map_lit_alias
package map_lit_alias

// A map LITERAL's keys must be copied at insert too, not just keys written via
// m[k] = v. Without resolving the key's children before the copy, an
// unresolved RefValue child is copied as a ref and then filled to point at the
// shared store object, so the stored key aliases the source variable and the
// entry becomes findable by a key it no longer equals.

type inner struct {
	n int
}

type key struct {
	a  int
	in inner
}

var (
	k key
	m map[key]int
)

func init() {
	k = key{a: 1, in: inner{n: 2}}
}

func main(cur realm) {
	m = map[key]int{k: 5}
	k.in.n = 99 // mutate the source AFTER the literal captured it
	for sk, sv := range m {
		println("stored key:", sk.a, sk.in.n, "value:", sv)
	}
}

// Output:
// stored key: 1 2 value: 5
