// Call-time dynamic dispatch facets. Reassigning an embedded interface field to
// a *different concrete type* re-dispatches at the call (redispatch 2). A
// pointer-receiver method promoted through an embedded field of a nil pointer
// panics cleanly instead of crashing, and the recovered value satisfies error
// like Go's runtime.Error (c5 2). And re-boxing through a second
// interface, nested unwrap, a value stored in a map, and one via type-assertion
// all observe the operand's current value at the call (rebox 9, nested 9,
// map 8, assert 9).

package main

type Deep struct{ x int }

func (d Deep) Get() int  { return d.x }
func (d *Deep) Ptr() int { return d.x }

type IG interface{ Get() int }
type IP interface{ Ptr() int }

type Impl struct{ x int }

func (i Impl) Get() int { return i.x }

type Other struct{}

func (Other) Get() int { return 2 }

type Box struct{ IG } // embeds the interface

type MidP struct{ Deep } // promotes Ptr (pointer receiver)

var pmp *MidP

// reassign the embedded interface field to a different concrete type.
func redispatch() int {
	s := &Box{IG: Impl{x: 1}}
	var o IG = s
	g := o.Get
	s.IG = Other{}
	return g()
}

// pointer-receiver method promoted through an embedded field of a nil pointer.
func c5() (r int) {
	defer func() {
		if _, ok := recover().(error); ok {
			r++
		}
	}()
	var i IP = pmp // nil *MidP
	defer i.Ptr()
	r = 1
	return
}

func reboxChain() int {
	var i1 IG = &Deep{x: 1}
	var i2 IG = i1 // re-box into another interface
	g := i2.Get
	(i1.(*Deep)).x = 9
	return g()
}

func nestedUnwrap() int {
	p := &Deep{x: 1}
	var i IG = &Box{IG: p} // *Box -> embedded IG -> *Deep -> Deep.Get
	g := i.Get
	p.x = 9
	return g()
}

func storedInMap() int {
	m := map[string]IG{"a": &Deep{x: 3}}
	g := m["a"].Get
	(m["a"].(*Deep)).x = 8
	return g()
}

func viaTypeAssert() int {
	var a interface{} = &Deep{x: 5}
	g := a.(IG).Get
	(a.(*Deep)).x = 9
	return g()
}

func main() {
	println("redispatch", redispatch())
	println("c5", c5())
	println("rebox", reboxChain())
	println("nested", nestedUnwrap())
	println("map", storedInMap())
	println("assert", viaTypeAssert())
}

// Output:
// redispatch 2
// c5 2
// rebox 9
// nested 9
// map 8
// assert 9
