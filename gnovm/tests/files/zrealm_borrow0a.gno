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

	crossrealm.SetObject(cross2(cur), p) // saved in crossrealm.

	p.Mutate()
}

// Note: 1. p and s are stored in `crossrealm`, so no borrow to `borrow_test`.
//       2. the `CurrentRealm` and `PreviousRealm` are `borrow_test`
//          and the origin user realm. which means the borrowed realm
//          is excluded from CurrentRealm and PreviousRealm.

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): s<VPBlock(1,0)>.A
