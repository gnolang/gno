// Method promotion through embedded fields. Through an interface the receiver
// is materialized at the call, so a post-bind mutation is seen — for an
// embedded value field (valueField 9) and an embedded *pointer field
// (ptrField 9). A method promoted from an embedded *interface* field dispatches
// dynamically through that field (ifaceField 7). But an embedded *pointer field
// with NO interface in the trail is static dispatch — snapshot at bind, so the
// mutation is invisible (ptrNoIface 1).

package main

type T struct{ x int }

func (t T) Get() int { return t.x }

type Mid struct{ T }   // value embed
type Top struct{ Mid } // two-level value embed
type PEmb struct{ *T } // embedded pointer field

type IG interface{ Get() int }

type Impl struct{ x int }

func (i Impl) Get() int { return i.x }

type Holder struct{ IG } // embeds the interface

func main() {
	// two-level embedded value field, via interface → mutation seen at call.
	t := &Top{Mid{T{x: 1}}}
	var i1 IG = t
	g1 := i1.Get
	t.x = 9
	println("valueField", g1())

	// embedded *pointer field, via interface → mutation seen at call.
	d := &T{x: 1}
	var i2 IG = &PEmb{d}
	g2 := i2.Get
	d.x = 9
	println("ptrField", g2())

	// embedded *interface* field → dynamic dispatch through the field.
	var i3 IG = &Holder{IG: Impl{x: 7}}
	g3 := i3.Get
	println("ifaceField", g3())

	// embedded *pointer field, NO interface → static dispatch, snapshot at bind.
	s := PEmb{T: &T{x: 1}}
	g4 := s.Get
	s.T.x = 9
	println("ptrNoIface", g4())
}

// Output:
// valueField 9
// ptrField 9
// ifaceField 7
// ptrNoIface 1
