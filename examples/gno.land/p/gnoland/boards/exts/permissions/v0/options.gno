package permissions

import (
	"strings"

	"gno.land/p/gnoland/boards/v0"
)

// Option configures permissions.
// It is a type with a method, not a func: a func value from this
// library runs view-only and could not write the new Permissions.
type Option interface {
	apply(*Permissions)
}

type useSingleUserRole struct{}

func (useSingleUserRole) apply(p *Permissions) { p.singleUserRole = true }

type withSuperRole struct{ r boards.Role }

func (o withSuperRole) apply(p *Permissions) {
	if p.superRole != "" {
		panic("permissions super role can be assigned only once")
	}

	name := string(o.r)
	if strings.TrimSpace(name) == "" {
		panic("permissions super role name is required")
	}

	if _, err := p.group.AddRole(name); err != nil {
		panic(err)
	}
	p.superRole = o.r
}

// UseSingleUserRole configures permissions to only allow one role per user.
func UseSingleUserRole() Option { return useSingleUserRole{} }

// WithSuperRole configures permissions to have a super role.
// A super role is the one that have all permissions.
// This type of role doesn't need to be mapped to any permission.
func WithSuperRole(r boards.Role) Option { return withSuperRole{r} }
