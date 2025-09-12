// PKGPATH: gno.land/r/borrow
package borrow

import crossrealm "gno.land/r/tests/vm/crossrealm_b"

func main() {
	crossrealm.Passthrough()
	crossrealm.Render()
}

// Error:
// invalid call: MockMutator<~VPBlock(3,35)>.Mutate(), method realm: RID46279D1E03ECB38B84822F9CAF584BD16913470D does not match receiver realm: RIDA7E973C6AEAB1F1CB747D0903CB1E4D369226376
