// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	tests "gno.land/r/tests/vm"
)

// Pointer-typed foreign var initialized via a foreign constructor
// that returns &CompositeLit. Distinct from crossrealm4 in that the
// pointer's target is a freshly heap-allocated HIV (not a slot in
// /r/tests/vm's package block):
//
//   - The constructor's composite literal allocates a *StructValue
//     stamped /r/tests/vm (borrow rule #1 makes
//     checkConstructionTime accept the literal even though the
//     caller is /r/crossrealm).
//   - `&CompositeLit` wraps in a HIV stamped with the active
//     currentRealmID — which is /r/tests/vm via borrow rule #1, not
//     /r/crossrealm.
//   - The returned pointer's Base is that HIV, NOT a block.
//
// At method-call time, IsReadonlyBy's HIV branch (ownership.go:466)
// recurses into the wrapped StructValue's authority, which matches
// m.Realm under borrow rule #2. Write succeeds.
var somevalue *tests.TestRealmObject

func init() {
	somevalue = tests.NewTestRealmObject()
}

func main(cur realm) {
	somevalue.Modify()
	println(somevalue)
}

// Output:
// &(struct{("initial_modified" string)} gno.land/r/tests/vm.TestRealmObject)
