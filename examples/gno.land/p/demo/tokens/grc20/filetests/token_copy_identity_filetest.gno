// PKGPATH: gno.land/r/demo/grc20copy

// A Token copied into a struct field or an array element has no object of its
// own: a pointer to it resolves to the container, which every sibling shares,
// and a slice resolves to its backing array. Rather than hand two ledgers one
// address, ObjectAddress reads "" for such a value. Only the pointer NewToken returns, or a copy held in its own
// variable, carries an id.
package grc20copy

import (
	"chain/runtime"

	"gno.land/p/demo/tokens/grc20"
)

type Pair struct {
	A grc20.Token
	B grc20.Token
}

var (
	origA, origB *grc20.Token
	pair         Pair
	arr          [2]grc20.Token
	sl           []grc20.Token
	varA, varB   grc20.Token
)

func init(cur realm) {
	origA, _ = grc20.NewToken("Alpha", "ALPHA", 4, cur)
	origB, _ = grc20.NewToken("Beta", "BETA", 4, cur)

	pair.A, pair.B = *origA, *origB
	arr[0], arr[1] = *origA, *origB
	sl = []grc20.Token{*origA, *origB}
	varA, varB = *origA, *origB
}

func main(cur realm) {
	println("pointers differ:", origA.ID() != origB.ID())
	println("separate vars differ:", varA.ID() != varB.ID())
	println("struct field id:", pair.A.ID() == "", pair.B.ID() == "")
	println("array element id:", arr[0].ID() == "", arr[1].ID() == "")
	println("slice element id:", sl[0].ID() == "", sl[1].ID() == "")
	println("slice id:", runtime.ObjectAddress(sl) == "")
}

// Output:
// pointers differ: true
// separate vars differ: true
// struct field id: true true
// array element id: true true
// slice element id: true true
// slice id: true
