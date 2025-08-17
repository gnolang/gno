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
// should not happen, mismatched method/receiver realm, MockMutator<~VPBlock(3,35)>.Mutate()/(gno.land/r/borrow---------RID1712AC7ADCFDC8E58A67E5615E20FB312394C4DF)
