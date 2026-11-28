// Cycle routed through a boxed struct value rather than a bare pointer: after
// the first hop the operand is the *StructValue W (not a PointerValue), so
// detection must key the visited set on struct identity too — an interface
// field re-read yields the same *StructValue each hop. Fatal, gno-unrecoverable
// (Go fatally stack-overflows). See method_iface_cyclic.gno.
package main

type IG interface{ Get() int }

type Box struct{ IG }

type W struct{ *Box }

func main() {
	s := &Box{}
	s.IG = W{s} // cycle through a value, not a bare pointer
	var o IG = s
	g := o.Get
	println(g())
}

// Error:
// cyclic embedded interface in method-value dispatch
