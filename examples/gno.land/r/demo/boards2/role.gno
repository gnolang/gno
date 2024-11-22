package boards2

// TODO: Rename file to "auth.gno" and define a new interface

type Permission string

const (
	PermissionEdit   Permission = "edit"
	PermissionDelete Permission = "delete"
)
