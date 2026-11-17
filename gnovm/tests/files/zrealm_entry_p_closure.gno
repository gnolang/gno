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
	println("owner literal:", revive(func() { entryclosure.LitIncer()() }))
	entryclosure.OwnerInc(cross(cur))
	println("owner entry wrote:", entryclosure.N(), entryclosure.Count())
}

// Output:
// p closure: cannot modify object owned by another realm from library code (call a function of the owning realm): b<~VPBlock(1,0)>.N
// p counter: cannot modify object owned by another realm from library code (call a function of the owning realm): c<~VPBlock(1,1)>
// n: 0 count: 0
// owner literal: gno.land/r/tests/vm/entryclosure's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/entryclosure) package{})).box.N
// owner entry wrote: 1 1
