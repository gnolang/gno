package access

const (
	errConfigNil      = "config cannot be nil"
	errNotInitialized = "access control not initialized"
	errEmptyRole      = "role name cannot be empty"
	errRoleExists     = "role %s already exists"
	errDeclareRole    = "failed to declare role %s: %v"
	errUnauthorized   = "caller(%s) is not authorized for role %s"
)
