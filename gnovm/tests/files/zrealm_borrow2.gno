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

func init() {
	s = &Struct{A: 100} // saved in borrow_test.
}

func main() {
	crossing()

	// cross(crossrealm_b.SetObject)(s) // saved in crossrealm_b

	p := &Passthrough{S: s}

	cross(crossrealm.SetObject)(p) // saved in crossrealm.

	p.Mutate() // receiver method borrows crossrealm_b.

	println(p.Get())

	p.PrintRealms()
}

// Output:
// 101
// CodeRealm{ g1evmgnh6rmk8vgahd4p2pc3reh0dw26twaz3rqq, gno.land/r/borrow_test }
// UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
