// What the lazy bind captures at the first interface crossing decides when a
// later change shows. The operand is captured at the crossing; navigation
// before it (including a concrete pointer deref) is eager.
//
//   - concrete root (value Box{} or pointer &Box): the embedded interface field
//     is read at bind, so reassigning it is invisible (valueRoot 1, ptrRoot 1).
//   - interface root (var g Getter = &Box): the operand is the *Box pointer, so
//     the field is re-read at the call → reassignment seen (ifaceRoot 9).
//   - mid interface holding a *pointer*: the captured operand aliases the
//     pointee, so mutating the pointee (not reassigning the field) is seen
//     (midPtrPointee 9).

package main

type Getter interface{ Get() int }

type Impl struct{ x int }

func (i Impl) Get() int { return i.x }

type Box struct{ Getter } // embeds the interface

type Deep struct{ x int }

func (d Deep) Get() int { return d.x }

func main() {
	// value root: embedded interface read at bind → reassignment invisible.
	bv := Box{Getter: Impl{x: 1}}
	gv := bv.Get
	bv.Getter = Impl{x: 9}
	println("valueRoot", gv())

	// concrete pointer root: still concrete → read at bind too.
	bp := &Box{Getter: Impl{x: 1}}
	gp := bp.Get
	bp.Getter = Impl{x: 9}
	println("ptrRoot", gp())

	// interface root: operand is *Box → field re-read at the call.
	bi := &Box{Getter: Impl{x: 1}}
	var g Getter = bi
	gi := g.Get
	bi.Getter = Impl{x: 9}
	println("ifaceRoot", gi())

	// mid interface holding a pointer: mutate the pointee (not the field) → seen.
	d := &Deep{x: 1}
	bm := Box{Getter: d}
	gm := bm.Get
	d.x = 9
	println("midPtrPointee", gm())
}

// Output:
// valueRoot 1
// ptrRoot 1
// ifaceRoot 9
// midPtrPointee 9
