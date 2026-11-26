// PKGPATH: gno.land/r/tests/vm/recoverfvrunbmv
package recoverfvrunbmv

import (
	"gno.land/p/demo/tests/recoverfv"
	"gno.land/r/tests/vm/recoverfvb"
)

// A method value on realm B's stored object runs inside a library-stored
// closure and recovers (master aborts); the same route cannot write B.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	ops := []func(func() any) any{recoverfv.StoredRun}
	println("rec:", ops[0](recoverfvb.H.Try))
	println("after")
	println("write:", try(func() { ops[0](recoverfvb.H.Inc) }))
	println("b state:", recoverfvb.H.K())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// after
// write: cannot modify object owned by another realm from library code (call a function of the owning realm): h<VPBlock(1,0)>.k
// b state: 0
// OUTCOME caught
