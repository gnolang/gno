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

func (s *Struct) PrintRealms() {
	s.printRealms()
}

func (s *Struct) printRealms() {
	// XXX
	// printRealm is called from s.PrintRealm()
	// which in this test is persisted in the same realm.
	// Therefore p.printRealm() -> s.PrintRealm()
	// borrows this realm, and then the crossing()
	// below is called.
	// Should this be allowed?
	// Whether it panics or not is determined by
	// 'fr.DidCrossing = true' in PushFrameCall.
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
	p.S.PrintRealms()
}

var s *Struct

func init() {
	s = &Struct{A: 100} // saved in borrow_test.
}

func main(cur realm) {
	// cross(crossrealm_b.SetObject)(s) // saved in crossrealm_b

	p := &Passthrough{S: s}

	crossrealm.SetObject(cross, p) // saved in crossrealm.

	p.Mutate() // receiver method borrows crossrealm_b.

	println(p.Get())

	p.PrintRealms()
}

// Output:
// 101
// CodeRealm{ g1mdevzlw3defhhew53zsky80lhqnk94ew8grqdt, gno.land/r/borrow }
// UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
