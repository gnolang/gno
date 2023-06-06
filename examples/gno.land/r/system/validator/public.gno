// This package is used to manage the validator set.
package validator

var (
	activeSet Set
	adminAddr std.Address
)

func ListActiveNodes() []Node {
	panic("not implemented")
}

func GetByOwnerAddr(ownerAddr std.Address) Node {
	panic("not implemented")
}

func GetByNodeAddr(nodeAddr std.Address) Node {
	panic("not implemented")
}

func RegisterNode(nodeAddr std.Address) (validatorID int) {
	// TODO: assertPaidFees (for spam prevention)
	n := Node{
		Name:      name,
		OwnerAddr: std.OrigAddr(),
		NodeAddr:  nodeAddr,
		IsActive:  false,
		ID:        idx,
	}
	set = append(set, n)
	idx++
}

func OwnerUpdateNodeAddr(addr std.Address) {
	// assertIsOwner
	panic("not implemented")
}

func AdminSetActiveStatus(ownerAddr std.Address, activeStatus bool) {
	// assertIsAdmin
	panic("not implemented")
}

func AdminUpdateAdminAddr(newAddr std.Address) {
	panic("not implemented")
}

// for later, maybe: slash, migrate, etc
