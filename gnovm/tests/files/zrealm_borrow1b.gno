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
	s.printRealms(cross2(cur))
}

func (s *Struct) printRealms(cur realm) {
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

func (p *Passthrough) PrintRealms(_ int, rlm realm) {
	p.printRealm(0, rlm)
}

func (p *Passthrough) printRealm(_ int, rlm realm) {
	p.S.PrintRealms(cross2(rlm))
}

var s *Struct

func init() {
	s = &Struct{A: 100} // saved in borrow_test.
}

func main(cur realm) {
	p := &Passthrough{S: s}

	crossrealm.SetObject(cross2(cur), p) // saved in crossrealm.

	p.Mutate() // receiver method borrows crossrealm.

	println(p.Get())

	p.PrintRealms(0, cur)
}

// Output:
// 101
// CodeRealm{ g1mdevzlw3defhhew53zsky80lhqnk94ew8grqdt, gno.land/r/borrow }
// CodeRealm{ g1mdevzlw3defhhew53zsky80lhqnk94ew8grqdt, gno.land/r/borrow }
