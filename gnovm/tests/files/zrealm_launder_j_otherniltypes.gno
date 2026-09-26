// PKGPATH: gno.land/r/launderj
package launderj

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Attack J: NIL-VALUED RECEIVERS OF OTHER REFERENCE-UNDERLYING TYPES.
//
// Variants of Attack H/I where the receiver is a nil function-typed
// value, nil slice, or nil map of a defined type. Each has
// `GetFirstObject` returning nil — same borrow rule #2 anchor gap. Each
// call is recovered separately so all three shapes are pinned.
//
// Same preconditions as Attack H: exploitable only when victim
// logic data is /p/-declared.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	gp := laundervictim.GetPtr()
	imm := laundervictim.GetImm()

	// One recover per shape, so every variant is pinned as refused.
	try := func(name string, m launderpkg.Mutator) {
		defer func() { println(name+":", recover()) }()
		gp.UseMutator(imm, m)
	}
	var nilFn launderattack.EvilFunc
	try("EvilFunc", nilFn)
	var nilSlice launderattack.EvilSlice
	try("EvilSlice", nilSlice)
	var nilMap launderattack.EvilMap
	try("EvilMap", nilMap)
	println("after:", laundervictim.ReadImm())
}

// Output:
// EvilFunc: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,1)>.Field
// EvilSlice: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,1)>.Field
// EvilMap: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,1)>.Field
// after: original-imm
