// PKGPATH: gno.land/r/tests/vm/recoverfvrewrite
package recoverfvrewrite

import (
	"gno.land/p/demo/tests/recoverfv"
	"gno.land/r/tests/vm/recoverfvb"
)

// Library code running as this realm recovers a panic from B's read-only
// code, then this realm writes its own state: the write works, B's does not.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

var cnt int

func main(cur realm) {
	c := recoverfv.New()
	println("rec:", c.RunRec(recoverfvb.Saved()))
	println("write:", c.RunRec(recoverfvb.SavedInc()))
	println("b state:", recoverfvb.N())
	cnt++
	println("cnt:", cnt)
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// write: gno.land/r/tests/vm/recoverfvb's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/recoverfvb) package{})).n
// b state: 0
// cnt: 1
// OUTCOME caught
