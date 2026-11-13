// PKGPATH: gno.land/r/demo/viewmvarg
package viewmvarg

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/r/tests/vm/viewdefault"
)

// A method value called by another realm's code, or from inside another
// realm's borrowed frame, runs view-only: it writes no realm's objects,
// neither the binder's passed back as arguments nor the caller's own.

func main(cur realm) {
	h := viewdefault.GetHandle()
	h.Set("a", "1")
	// The owner's stdlib method value, the owner's view as its out-parameter.
	println("owner's arg:", revive(func() { viewdefault.Encoder()(viewdefault.Secret(), []byte("pwned!!!")) }))
	println("secret:", viewdefault.SecretString())
	// The caller's own tree from inside the owner's handle: refused as a
	// method value, written through a closure the caller declares.
	local := avl.NewTree()
	println("own tree, method value:", revive(func() { h.Iterate("", "", local.Set) }))
	h.Iterate("", "", func(k string, v any) bool { return local.Set(k, v) })
	println("own tree, closure:", local.Size())
}

// Output:
// owner's arg: cannot directly modify readonly tainted object (use a method or crossing function): dst<VPBlock(2,1)>[di<VPBlock(2,3)> + (const (0 int))]
// secret: secret-bytes
// own tree, method value: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
// own tree, closure: 2
