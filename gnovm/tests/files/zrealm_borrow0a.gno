// PKGPATH: gno.land/r/borrow
package borrow

import (
	"chain/runtime"
	"fmt"

	"gno.land/r/tests/vm/crossrealm"
)

type Struct struct {
	A int
}

func (s *Struct) Mutate() {
	s.A += 1
	fmt.Printf("s.A: %d\n", s.A)
	fmt.Println("current realm: ", runtime.CurrentRealm())
	fmt.Println("previous realm: ", runtime.PreviousRealm())
}

type Passthrough struct {
	S *Struct
}

func (p *Passthrough) Mutate() {
	p.S.Mutate()
}

func main(cur realm) {
	s := &Struct{A: 100}
	p := &Passthrough{S: s}

	crossrealm.SetObject(cross, p) // saved in crossrealm.

	p.Mutate()
}

// Note: 1. p and s are stored in `crossrealm`, so no borrow to `borrow_test`.
//       2. the `CurrentRealm` and `PreviousRealm` are `borrow_test`
//          and the origin user realm. which means the borrowed realm
//          is excluded from CurrentRealm and PreviousRealm.

// Output:
// s.A: 101
// current realm:  CodeRealm{ g1mdevzlw3defhhew53zsky80lhqnk94ew8grqdt, gno.land/r/borrow }
// previous realm:  UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
