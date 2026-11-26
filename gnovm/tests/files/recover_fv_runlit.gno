// PKGPATH: gno.land/r/tests/vm/recoverfvrunlit
package recoverfvrunlit

import "gno.land/p/demo/tests/recoverfv"

// A library-stored closure called by value from this realm's literal inside
// another library-stored closure recovers, as on master.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	ops := []func(func() any) any{recoverfv.StoredRun}
	inner := []func() any{recoverfv.StoredTry}
	println("rec:", ops[0](func() any { return inner[0]() }))
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// OUTCOME caught
