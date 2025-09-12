// PKGPATH: gno.land/r/borrow
package borrow

import (
	"gno.land/r/tests/vm/crossrealm"
)

type NoopMutator struct {
	num int
}

func (nm *NoopMutator) Mutate() {
	nm.num++
	println("only mutate local var, num: ", nm.num)
}

func main() {
	var nm = &NoopMutator{}
	crossrealm.SetMutator(cross, nm)
	crossrealm.DoMutate()
}

// Error:
// invalid call: MockMutator<~VPBlock(3,35)>.Mutate(), method realm: RIDC459C7103CD5303D426ED1E6E642FFFCFFD55F7C does not match receiver realm: RIDA7E973C6AEAB1F1CB747D0903CB1E4D369226376
