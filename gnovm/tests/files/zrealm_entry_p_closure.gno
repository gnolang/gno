// PKGPATH: gno.land/r/demo/entrypclosure
package entrypclosure

import "gno.land/r/tests/vm/entryclosure"

// A closure built by library code acts for no realm: called by a stranger it
// cannot write the owner's objects. A closure written in the owner's source,
// called without crossing, is read-only too; only the owner's crossing entry
// writes, through either kind.

func call(cur realm, f func()) { f() }

func main(cur realm) {
	println("p closure:", revive(func() { call(cross(cur), entryclosure.Incer()) }))
	println("p counter:", revive(func() { call(cross(cur), func() { entryclosure.Counter()() }) }))
	println("n:", entryclosure.N(), "count:", entryclosure.Count())
	println("owner literal:", try(func() { entryclosure.LitIncer()() }))
	entryclosure.OwnerInc(cross(cur))
	println("owner entry wrote:", entryclosure.N(), entryclosure.Count())
}

// try runs fn and returns what it panicked with: a refused write in a
// view-only frame is recovered by the caller, not revived.
func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): c<~VPBlock(1,1)>
