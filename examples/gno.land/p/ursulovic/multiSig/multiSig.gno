package multi_sig

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	owners            avl.Tree
	newOwnerApprovals avl.Tree
)

func init() {
}

func InitOwners(addresses []string) {
	if owners.Size() != 0 {
		panic("Owners list has already been initiated!")
	}

	for _, v := range addresses {
		addr := std.Address(v)
		if !addr.IsValid() {
			panic(ufmt.Sprintf("Entry %s is not a valid address!", v))
		}
		owners.Set(v, true)
	}
}

func AddNewOwner(address string) {
	caller := std.GetOrigCaller()

	if owners.Has(address) {
		panic("Proposed address is already an owner!")
	}

	if !owners.Has(caller.String()) {
		panic("Only owners can suggest new owners!")
	}

	v, exists := newOwnerApprovals.Get(address)
	if !exists {
		v = avl.NewTree()
		newOwnerApprovals.Set(address, v)
	}

	addressApprovals := v.(*avl.Tree)
	if addressApprovals.Has(caller.String()) {
		panic("You have already approved this new owner proposal address!")
	}
	addressApprovals.Set(caller.String(), true)

	if addressApprovals.Size() == owners.Size() {
		newOwnerApprovals.Remove(address)
		owners.Set(address, true)
	}

}
