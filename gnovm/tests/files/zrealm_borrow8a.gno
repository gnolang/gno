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
	crossrealm.SetDoMutate(cross, nm) // execute method of unreal receiver, cross to current realm
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(gno.land/r/demo/tests/crossrealm) package{})).AllowedList
