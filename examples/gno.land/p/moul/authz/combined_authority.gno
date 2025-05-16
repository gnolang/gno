package authz

/*
import (
	"std"
)

// CombinedAuthority implements Authority by combining multiple authorities.
// It grants permission if ANY of the underlying authorities grant permission.
// This is useful for combining different authorization strategies or for
// implementing complex permission hierarchies.
//
// Example:
//
//	adminAuth := New(adminAddr)
//	tokenAuth := NewTokenAuthority(tokenAddr)
//	combined := NewCombinedAuthority(adminAuth, tokenAuth)
//	if combined.Can(user, "action") {
//	    // user can perform action if either admin or token allows it
//	}
type CombinedAuthority struct {
	authorities []Authority
}

// NewCombinedAuthority creates a new CombinedAuthority from a list of authorities.
func NewCombinedAuthority(authorities ...Authority) *CombinedAuthority {
	return &CombinedAuthority{
		authorities: authorities,
	}
}

// Can checks if the caller is authorized to perform an action by checking
// all underlying authorities. Returns true if ANY authority grants permission.
func (c *CombinedAuthority) Can(addr std.Address, action string) bool {
	for _, auth := range c.authorities {
		if auth.MembersFor(action).Has(addr) {
			return true
		}
	}
	return false
}
*/
