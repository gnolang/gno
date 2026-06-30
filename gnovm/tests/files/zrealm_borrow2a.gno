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
	s = &Struct{A: 100}                   // saved in borrow_test.
	crossrealm_b.SetObject(cross(cur), s) // saved in crossrealm_b

	p := &Passthrough{S: s}

	crossrealm.SetObject(cross(cur), p) // saved in crossrealm.

	p.Mutate() // receiver method borrows crossrealm, then crossrealm_b, but does not impact prev, current realm.

	println(p.Get())
}

// Error:
// gno.land/r/borrow/zrealm_borrow2a.gno:16:26-31: use of builtin cross not in function call

// TypeCheckError:
// gno.land/r/borrow/zrealm_borrow2a.gno:16:26: cannot use cross (value of type func(rlm realm) realm) as crossrealm_b.realm value in argument to crossrealm_b.SetClosure: func(rlm realm) realm does not implement gno0p9.realm (missing method Address)
