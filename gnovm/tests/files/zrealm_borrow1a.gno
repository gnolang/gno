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
}

func (s *Struct) Get() int {
	return s.A
}

func (s *Struct) PrintRealms(cur realm) {
	s.printRealms()
}

func (s *Struct) printRealms() {
	fmt.Println(runtime.CurrentRealm())
	fmt.Println(runtime.PreviousRealm())
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
	p.S.PrintRealms(cross)
}

var s *Struct

func init() {
	s = &Struct{A: 100} // saved in borrow_test.
}

func main(cur realm) {
	p := &Passthrough{S: s}

	crossrealm.SetObject(cross, p) // saved in crossrealm.

	p.Mutate() // receiver method borrows crossrealm.

	println(p.Get())

	p.PrintRealms()
}

// Output:
// 101
// CodeRealm{ g1mdevzlw3defhhew53zsky80lhqnk94ew8grqdt, gno.land/r/borrow }
// CodeRealm{ g1mdevzlw3defhhew53zsky80lhqnk94ew8grqdt, gno.land/r/borrow }
