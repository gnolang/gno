// PKGPATH: gno.land/r/tests/vm/recoverfvrunmvlocal
package recoverfvrunmvlocal

import "gno.land/p/demo/tests/recoverfv"

// The same with a method value on a fresh local object (master aborts).

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	h := &recoverfv.H{}
	ops := []func(func() any) any{recoverfv.StoredRun}
	runInc := ops
	println("rec:", ops[0](h.Try))
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
