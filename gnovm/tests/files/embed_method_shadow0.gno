package main

// A method promoted from a shallower embedded field shadows a same-named
// method promoted from a deeper one (Go spec, "Selectors"). The struct
// therefore implements the interface, and selection resolves to the
// shallow method.

type I interface{ M() string }

type Deep struct{ V string }

func (Deep) M() string { return "deep" }

type Mid struct{ Deep } // M at depth 2

type Shallow struct{ V string }

func (Shallow) M() string { return "shallow" } // M at depth 1, shadows Mid.Deep.M

type T struct {
	Shallow
	Mid
}

func main() {
	t := T{Shallow{"shallowV"}, Mid{Deep{"deepV"}}}
	// interface satisfaction (was wrongly rejected at preprocess)
	var i I = t
	println("iface:", i.M())
	// direct selection resolves to the shallow method and field
	println("method:", t.M())
	println("field:", t.V)
}

// Output:
// iface: shallow
// method: shallow
// field: shallowV
