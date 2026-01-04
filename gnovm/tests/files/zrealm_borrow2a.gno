// PKGPATH: gno.land/r/borrow
package borrow

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

func (s *Struct) Mutate() {
	s.A += 1
	// modify object associated to borrowed realm
	crossrealm_b.SetClosure(cross, func() { println("foo closure") })
}

func (s *Struct) Get() int {
	return s.A
}

type Passthrough struct {
	S *Struct
}

func (p *Passthrough) Mutate() {
	p.S.Mutate()
}

func (p *Passthrough) Get() int {
	crossrealm_b.Closure()
	return p.S.Get()
}

var s *Struct

func main(cur realm) {
	s = &Struct{A: 100}              // saved in borrow_test.
	crossrealm_b.SetObject(cross, s) // saved in crossrealm_b

	p := &Passthrough{S: s}

	crossrealm.SetObject(cross, p) // saved in crossrealm.

	p.Mutate() // receiver method borrows crossrealm, then crossrealm_b, but does not impact prev, current realm.

	println(p.Get())
}

// Output:
// foo closure
// 101
