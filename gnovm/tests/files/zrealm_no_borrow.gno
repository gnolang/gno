// PKGPATH: gno.land/r/borrow_test
package borrow_test

import (
	"fmt"
	"std"

	"gno.land/r/demo/tests/crossrealm"
)

type Struct struct {
	A int
}

func (s *Struct) Mutate() {
	s.A += 1
	fmt.Printf("s.A: %d\n", s.A)
	fmt.Println("current realm: ", std.CurrentRealm())
	fmt.Println("previous realm: ", std.PreviousRealm())
}

type Passthrough struct {
	S *Struct
}

func (p *Passthrough) Mutate() {
	p.S.Mutate()
}

func main() {
	crossing()

	s := &Struct{A: 100}
	p := &Passthrough{S: s}

	cross(crossrealm.SetObject)(p) // saved in crossrealm.

	p.Mutate()
}

// Note: 1. p and s are stored in `crossrealm`, so no borrow to `borrow_test`.
//       2. the `CurrentRealm` and `PreviousRealm` are `borrow_test`
//          and the origin user realm. which means the borrowed realm
//          is excluded from CurrentRealm and PreviousRealm.

// Output:
// s.A: 101
// current realm:  CodeRealm{ g1evmgnh6rmk8vgahd4p2pc3reh0dw26twaz3rqq, gno.land/r/borrow_test }
// previous realm:  UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
