// PKGPATH: gno.land/r/tests/vm/recoverfvrunq
package recoverfvrunq

import (
	"gno.land/p/demo/tests/recoverfv"
	"gno.land/p/demo/tests/recoverfvq"
)

// Another library's stored closure runs inside a library-stored closure
// and recovers (master aborts).

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	ops := []func(func() any) any{recoverfv.StoredRun}
	runInc := ops
	println("rec:", ops[0](recoverfvq.StoredQ))
	println("after")
	println("write:", try(func() { runInc[0](func() any { recoverfv.StoredInc(); return nil }) }))
	println("lib state:", recoverfv.Shared().N())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// after
// write: cannot modify object owned by another realm from library code (call a function of the owning realm): (const (ref(gno.land/p/demo/tests/recoverfv) package{})).shared.n
// lib state: 0
// OUTCOME caught
