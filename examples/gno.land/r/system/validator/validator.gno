// This package is used to manage the validator set.
package validator

var (
	idx int = 0 // just a counter, not related to power
	adminAddr std.Address
	set Set
)

type Node struct {
	// owner fields
	Name     string
	NodeAddr std.Address

	// admin fields
	IsActive bool

	// internal fields
	ID        int
	OwnerAddr std.Address
}

type Set struct {
	Nodes []Node
	// room for additional set configuration
}

func ListActive() Set {
	panic("not implemented")
}

func Register(nodeAddr std.Address) (validatorID int) {
	// TODO: assertPaidFees (for spam prevention)
	n := Node{
		Name: name,
		OwnerAddr: std.OrigAddr(),
		NodeAddr: nodeAddr,
		IsActive: false,
		ID: idx,
	}
	set = append(set, n)
	idx++
}

func UpdateName(validatorID, name string) {
	// assertIsOwner
	panic("not implemented")
}

func UpdateNodeAddr(validatorID, addr std.Address) {
	// assertIsOwner
	panic("not implemented")
}

// AdminSetActive allows updating the active
func AdminSetActive(validatorID int, status bool) {
	// assertIsAdmin
	panic("not implemented")
}

func UpdateAdminAddr(newAddr std.Address) {
	panic("not implemented")
}


// for later, maybe: slash, migrate, etc
