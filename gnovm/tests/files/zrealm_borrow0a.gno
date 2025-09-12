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

func main(cur realm) {
	s := &Struct{A: 100}
	p := &Passthrough{S: s}

	crossrealm.SetObject(cross, p) // saved in crossrealm.

	p.Mutate() // cross into current realm where method Mutate is declared
}

// Error:
// invalid call: p<VPBlock(1,2)>.Mutate(), method realm: RIDC459C7103CD5303D426ED1E6E642FFFCFFD55F7C does not match receiver realm: RIDA7E973C6AEAB1F1CB747D0903CB1E4D369226376
