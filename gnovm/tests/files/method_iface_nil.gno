// Nil-receiver timing (matching Go). A concrete nil `*T` value method derefs
// when the method value is formed, so `defer pt.Get()` panics eagerly and r=1
// never runs (concrete 0). An interface-boxed nil defers the deref to the call,
// so recover catches it after r=1 ran — called immediately (ifaceImmediate 1)
// or promoted through an embedded interface (embedded 1). The deferred form is
// in method_nil_value_bind.gno (the initial failing-test artifact).

package main

type T struct{ x int }

func (t T) Get() int { return t.x }

type I interface{ Get() int }

type Mid struct{ T } // promotes Get

var pt *T
var pmid *Mid

// concrete nil: deref at bind → eager panic → r=1 never runs.
func concrete() (r int) {
	defer func() { recover() }()
	defer pt.Get()
	r = 1
	return
}

// interface nil, immediate call: deref at the call → recover after r=1.
func ifaceImmediate() (r int) {
	defer func() { recover() }()
	var i I = pt
	g := i.Get
	r = 1
	_ = g()
	r = 2 // unreached
	return
}

// embedded promotion through a nil pointer, via interface, deferred.
func embedded() (r int) {
	defer func() { recover() }()
	var i I = pmid // typed-nil *Mid, Get promoted
	defer i.Get()
	r = 1
	return
}

func main() {
	println("concrete", concrete())
	println("ifaceImmediate", ifaceImmediate())
	println("embedded", embedded())
}

// Output:
// concrete 0
// ifaceImmediate 1
// embedded 1
