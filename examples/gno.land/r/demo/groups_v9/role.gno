package groups

type Permission string

const (
	DeletePermission Permission = "role:delete"
	EditPermission   Permission = "role:edit"
)
