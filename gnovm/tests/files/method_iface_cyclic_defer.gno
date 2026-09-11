// A cyclic bind invoked via defer is resolved at the deferred-call site (the
// second resolveLazyBound caller), not the immediate one, so this covers that
// path. "before return" prints, then the deferred call fatally panics. See
// method_iface_cyclic.gno.
package main

type IG interface{ Get() int }

type Box struct{ IG }

func main() {
	s := &Box{}
	s.IG = s
	var o IG = s
	g := o.Get
	defer g()
	println("before return")
}

// Output:
// before return

// Error:
// cyclic embedded interface in method-value dispatch
