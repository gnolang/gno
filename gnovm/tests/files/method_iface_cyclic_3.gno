// Longer indirect cycle (a->b->c->a): detection must use a visited set, since no
// two consecutive operands repeat. Fatal, gno-unrecoverable (Go fatally
// stack-overflows). See method_iface_cyclic.gno.
package main

type IG interface{ Get() int }

type Box struct{ IG }

func main() {
	a, b, c := &Box{}, &Box{}, &Box{}
	a.IG = b
	b.IG = c
	c.IG = a
	var o IG = a
	g := o.Get
	println(g())
}

// Error:
// cyclic embedded interface in method-value dispatch
