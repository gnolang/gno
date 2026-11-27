// PKGPATH: gno.land/r/tests/vm/recoverfvbsaved
package recoverfvbsaved

import "gno.land/r/tests/vm/recoverfvb"

// A closure realm B made and saved, called by this realm, panics on B's
// object and is recovered here (master aborts); it cannot write B.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	println("rec:", try(recoverfvb.Saved()))
	println("write:", try(recoverfvb.SavedInc()))
	println("b state:", recoverfvb.N())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// write: gno.land/r/tests/vm/recoverfvb's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/recoverfvb) package{})).n
// b state: 0
// OUTCOME caught
