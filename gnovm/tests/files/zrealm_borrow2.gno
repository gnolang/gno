// PKGPATH: gno.land/r/borrow
package borrow

import (
	"fmt"
	"std"

	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

func (s *Struct) Mutate() {
	s.A += 1
}

func (s *Struct) Get() int {
	return s.A
}

func (s *Struct) PrintRealms() {
	s.printRealms()
}

func (s *Struct) printRealms() {
	fmt.Println(std.CurrentRealm())
	fmt.Println(std.PreviousRealm())
}

type Passthrough struct {
	S *Struct
}

func (p *Passthrough) Mutate() {
	p.S.Mutate()
}

func (p *Passthrough) Get() int {
	return p.S.Get()
}

func (p *Passthrough) PrintRealms() {
	p.printRealm()
}

func (p *Passthrough) printRealm() {
	p.S.PrintRealms()
}

var s *Struct

func main(cur realm) {
	s = &Struct{A: 100}              // saved in current realm.
	crossrealm_b.SetObject(cross, s) // saved in crossrealm_b

	p := &Passthrough{S: s}

	crossrealm.SetObject(cross, p) // saved in crossrealm.

	p.Mutate() // cross into current realm where method Mutate is declared

	println(p.Get())

	p.PrintRealms()
}

// Error:
// cannot directly modify readonly tainted object (w/o method): s<VPBlock(1,0)>.A
