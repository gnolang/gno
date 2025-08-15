// PKGPATH: gno.land/r/borrow
package borrow

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

var s = &Struct{A: 100}
var p *Passthrough

func main(cur realm) {
	p = &Passthrough{S: s}

	crossrealm.SetObject(cross, p)

	p.Mutate() // cross into current realm where method Mutate is declared
}

// Error:
// should not happen, method and object not in same realm
