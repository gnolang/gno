package permissions

import (
	"gno.land/p/gnoland/boards"
)

// Option configures permissions.
type Option func(*Permissions)

// UseSingleUserRole configures permissions to only allow one role per user.
func UseSingleUserRole() Option {
	return func(p *Permissions) {
		p.singleUserRole = true
	}
}

// WithSuperRole configures permissions to have a super role.
// A super role is the one that have all permissions.
// This type of role doesn't need to be mapped to any permission.
func WithSuperRole(r boards.Role) Option {
	return func(p *Permissions) {
		if p.superRole != "" {
			panic("permissions super role can be assigned only once")
		}

		name := string(r)
		p.dao.Members().Grouping().Add(name)
		p.superRole = r
	}
}
