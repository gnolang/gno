// The same interface method set must have one TypeID regardless of how it was
// built: an unexported method listed directly (PkgPath left unstamped by the
// no-embed fast path) vs promoted through an embed (PkgPath stamped with the
// origin package). The TypeID sort and emission must use the same package
// fallback, else the unstamped and stamped forms diverge. Guards the
// construction-level half of the sort/emit provenance fix (PR #5739 review);
// the persisted-decode half is covered by TestInterfaceTypeID_PkgPathProvenance.
package main

type I = interface {
	M() int
	z() int
}

func main() {
	var a, b interface{} = struct {
		X interface {
			M() int
			z() int
		}
	}{}, struct{ X interface{ I } }{}
	println("direct==embedded:", a == b)
}

// Output:
// direct==embedded: true
