// PKGPATH: gno.land/r/borrow
package borrow

import (
	"gno.land/r/demo/tests/crossrealm"
)

type NoopMutator struct{}

func (nm *NoopMutator) Mutate() {
	crossrealm.AllowedList = append(crossrealm.AllowedList, 3)
	println(crossrealm.AllowedList)
}

func main() {
	var nm = &NoopMutator{}
	crossrealm.SetMutator(cross, nm)
	crossrealm.DoMutate()
}

// Error:
// should not happen, method and object not in same realm
