// Indirect (a->b->a) cyclic embedded interface: the dispatch loop never sees an
// adjacent repeat, only a repeat after one hop, so detection must use a visited
// set (not a last-operand check). Raises the same fatal, gno-unrecoverable panic
// as the direct variant. See method_iface_cyclic.gno.
package main

type IG interface{ Get() int }

type Box struct{ IG }

func main() {
	defer func() { println("recover:", recover()) }() // does NOT catch a fatal panic
	a := &Box{}
	b := &Box{}
	a.IG = b
	b.IG = a // indirect 2-cycle: a -> b -> a
	var o IG = a
	g := o.Get
	println(g())
}

// Error:
// cyclic embedded interface in method-value dispatch
