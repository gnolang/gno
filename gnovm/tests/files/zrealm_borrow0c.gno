// PKGPATH: gno.land/r/borrow
package borrow

import (
	"fmt"
	"std"

	"gno.land/r/tests/vm/crossrealm"
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

var s = &Struct{A: 100}
var p = &Passthrough{S: s}

func main(cur realm) {
	crossrealm.SetObject(cross, p) // p already attached to current realm

	p.Mutate() // cross into current realm where method Mutate is declared
}

// Output:
// s.A: 101
// current realm:  CodeRealm{ g1mdevzlw3defhhew53zsky80lhqnk94ew8grqdt, gno.land/r/borrow }
// previous realm:  UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
