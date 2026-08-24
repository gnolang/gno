// A cyclic embedded interface makes interface method-value dispatch
// non-convergent: resolveLazyBound would re-box the same operand forever. It
// tracks the operand pointer and raises a fatal, gno-unrecoverable panic when
// one repeats — matching Go, which runs the same program as recursion and
// fatally stack-overflows. recover() cannot catch it (see indirect variant for
// the a->b->a case, which exercises the seen-set rather than an adjacent check).
package main

type IG interface{ Get() int }

type Box struct{ IG }

func main() {
	defer func() { println("recover:", recover()) }() // does NOT catch a fatal panic
	s := &Box{}
	s.IG = s // direct: the embedded interface points back at the box
	var o IG = s
	g := o.Get
	println(g())
}

// Error:
// cyclic embedded interface in method-value dispatch
