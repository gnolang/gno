package validator

type Node struct {
	NodeAddr  std.Address
	OwnerAddr std.Address
	// Statistics
}

func (n Node) Name() string {
	// lookup r/users/names?
	panic("not implemented")
}
