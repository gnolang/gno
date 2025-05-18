package authz

type AccessControl interface {
	Can(addr std.Address, action string, resource any) bool
}
