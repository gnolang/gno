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
// cannot directly modify readonly tainted object (w/o method): nm<VPBlock(1,0)>.num
