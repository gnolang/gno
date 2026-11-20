// PKGPATH: gno.land/r/launderrh_mv_returned_closure
package launderrh_mv_returned_closure

// /r/-victim returns a closure that writes gImm. Called by the attacker
// without crossing, the victim's code is read-only, so the write is refused.
// A crossing closure the victim returns, called with cross, runs as the victim
// and writes: returning it is the victim's consent.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	w := launderrvictim.MakeWriterClosure("pwnd-via-returned-closure")
	println("plain closure:", try(w))
	println("after:", launderrvictim.ReadImm())

	wx := launderrvictim.MakeWriterX("written-via-crossing-closure")
	wx(cross(cur))
	println("after crossing closure:", launderrvictim.ReadImm())
}

// try runs fn and returns what it panicked with: a refused write in a
// view-only frame is recovered by the caller, not revived.
func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// before: rdata-original
// plain closure: gno.land/r/tests/vm/launderrvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/launderrvictim) package{})).gImm.Field
// after: rdata-original
// after crossing closure: written-via-crossing-closure
