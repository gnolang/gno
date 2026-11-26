// PKGPATH: gno.land/r/decompose_iface
package decompose_iface

import "gno.land/r/tests/vm/launderrvictim"

// Assigning into an interface, named or not, boxes the result: no conversion,
// so a foreign realm's pointer can be held as it can with `b = x`.

type Box interface{}

type Stringer interface{ Read() string }

type T struct{}

func (T) Read() string { return "T" }

func local() (*T, bool) { return &T{}, true }

func foreign() (*launderrvictim.Immutable, bool) { return launderrvictim.GetImm(), true }

func main(cur realm) {
	var b Box
	var s Stringer
	var ok bool
	b, ok = local()
	_, isT := b.(*T)
	println(isT, ok)
	s, ok = local()
	println(s.Read(), ok)
	b, ok = foreign()
	_, isImm := b.(*launderrvictim.Immutable)
	println(isImm, ok)
	s, ok = foreign()
	println(s.Read(), ok)
}

// Output:
// true true
// T true
// true true
// rdata-original true
