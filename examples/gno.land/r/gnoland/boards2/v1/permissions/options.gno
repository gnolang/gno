package permissions

// Option configures permissions.
type Option func(*Permissions)

// UseSingleUserRole configures permissions to only allow one role per user.
func UseSingleUserRole() Option {
	return func(p *Permissions) {
		p.singleUserRole = true
	}
}
