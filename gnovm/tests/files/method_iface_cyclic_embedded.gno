// Cycle through a two-level embed: *Box satisfies IG via Box->Inner->IG
// promotion, and the embedded interface points back at the box. The dispatch
// walk strips the promoted layers and still loops, so detection fires. Fatal
// (Go fatally stack-overflows). See method_iface_cyclic.gno.
package main

type IG interface{ Get() int }

type Inner struct{ IG }

type Box struct{ Inner }

func main() {
	s := &Box{}
	s.Inner.IG = s
	var o IG = s
	g := o.Get
	println(g())
}

// Error:
// cyclic embedded interface in method-value dispatch
