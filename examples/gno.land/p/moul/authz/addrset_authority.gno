package authz

/*
import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/addrset"
)

// AddrsetAuthority is an opinionated implementation of the Authority interface
// that uses address sets to manage permissions for different actions.
//
// It uses a two-layer design:
//   - AddrsetAuthority: A safe, public interface that enforces admin-only
//     policy modifications. It implements the Authority interface and provides
//     methods for checking permissions (Can) and managing policies (Add, Remove,
//     RemoveAll) that can only be called by administrators.
//   - AddrsetPrivateConfig: An internal, unrestricted configuration object that
//     allows direct policy modifications. It's returned by New() to enable the
//     realm to set up initial policies, but should not be exposed to external
//     callers.
//
// Usage example:
//
//	auth, privcfg := authz.New(adminAddr)
//	privcfg.Add("propose", user1, user2)  // Set up initial policies
//	if auth.Can(user1, "propose") {       // Check permissions
//	    // user1 can propose
//	}
//
// Special addresses:
//   - Anyone: A special address that grants permission to any caller
//   - Admin: A special action that grants full administrative privileges
//
// This implementation is designed to be simple to use while providing enough
// flexibility for most authorization needs in Gno realms.
type AddrsetAuthority struct {
	config AddrsetPrivateConfig
}

// AddrsetPrivateConfig is the private configuration of the AddrsetAuthority.
// It isn't a safe object and is intended to be kept unexported.
type AddrsetPrivateConfig struct {
	policies avl.Tree // action -> addrset.Set
}

// New creates a new empty AddrsetAuthority and its AddrsetPrivateConfig.
func New(admins ...std.Address) (*AddrsetAuthority, *AddrsetPrivateConfig) {
	auth := &AddrsetAuthority{}
	for _, addr := range admins {
		auth.config.Add(Admin, addr)
	}
	return auth, &auth.config
}

// NewWithCurrent creates a new Authority with the current realm as an admin.
func NewWithCurrent() (*AddrsetAuthority, *AddrsetPrivateConfig) {
	return New(std.CurrentRealm().Address())
}

// NewWithPrevious creates a new Authority with the previous realm as an admin.
func NewWithPrevious() (*AddrsetAuthority, *AddrsetPrivateConfig) {
	return New(std.PreviousRealm().Address())
}

const (
	// Anyone is a special address that can be used to approve any caller
	// address.
	Anyone = std.Address("g1anyone")

	// Admin is a special action that can be used as the action name for a
	// policy allowing the administrators to do any action.
	// It's used to make the AddrsetAuthority implementation a safe object,
	// allowing any administrator to manage the policies if the object is
	// exposed.
	Admin = "_admin_"
)

// AddrsetAuthority implementation
//

// Can checks if the caller is authorized to perform an action.
func (a *AddrsetAuthority) Can(addr std.Address, action string) bool {
	for _, lookup := range []string{action, Admin} {
		value, ok := a.config.policies.Get(lookup)
		if !ok {
			continue
		}
		addrs := value.(*addrset.Set)
		if addrs.Has(addr) || addrs.Has(Anyone) {
			return true
		}
	}
	return false
}

// Add adds members to a policy, but only if called by an admin.
func (a *AddrsetAuthority) Add(action string, members ...std.Address) {
	if !a.Can(std.CurrentRealm().Address(), Admin) {
		panic("unauthorized: only admin can add members")
	}
	a.config.Add(action, members...)
}

// Remove removes members from a policy, but only if called by an admin.
func (a *AddrsetAuthority) Remove(action string, members ...std.Address) {
	if !a.Can(std.CurrentRealm().Address(), Admin) {
		panic("unauthorized: only admin can remove members")
	}
	a.config.Remove(action, members...)
}

// RemoveAll removes all policies for a given action, but only if called by an admin.
func (a *AddrsetAuthority) RemoveAll(action string) {
	if !a.Can(std.CurrentRealm().Address(), Admin) {
		panic("unauthorized: only admin can remove all policies")
	}
	a.config.RemoveAll(action)
}

// String returns a string representation of the AddrsetAuthority.
func (a *AddrsetAuthority) String() string {
	return a.config.String()
}

// AddrsetPrivateConfig implementation
//

// Add adds members to a policy.
func (c *AddrsetPrivateConfig) Add(action string, members ...std.Address) {
	value, ok := c.policies.Get(action)
	if !ok {
		value = &addrset.Set{}
		c.policies.Set(action, value)
	}
	addrs := value.(*addrset.Set)
	for _, member := range members {
		addrs.Add(member)
	}
}

// Remove removes members from a policy.
func (c *AddrsetPrivateConfig) Remove(action string, members ...std.Address) {
	value, ok := c.policies.Get(action)
	if !ok {
		return
	}
	addrs := value.(*addrset.Set)
	for _, member := range members {
		addrs.Remove(member)
	}
	if addrs.Size() == 0 {
		c.policies.Remove(action)
	}
}

// RemoveAll removes all policies for a given action.
func (c *AddrsetPrivateConfig) RemoveAll(action string) {
	c.policies.Remove(action)
}

// String returns a string representation of the AddrsetPrivateConfig.
// It limits the number of entries to 5 and shows the remaining entries if there
// are more.
// It also shows the first two addresses for each policy and the remaining
// addresses if there are more.
func (c *AddrsetPrivateConfig) String() string {
	var result string
	count := 0
	total := c.policies.Size()
	c.policies.Iterate("", "", func(key string, value any) bool {
		action := key
		if count >= 5 {
			remaining := total - count
			if result != "" {
				result += "\n"
			}
			result += ufmt.Sprintf("(%d more entries...)", remaining)
			return true // stop iteration
		}
		addrs := value.(*addrset.Set)
		if result != "" {
			result += "\n"
		}
		addrSize := addrs.Size()
		switch {
		case addrSize == 0:
			result += action + ": (empty)"
		case addrSize < 10:
			result += action + ": "
			addrs.IterateByOffset(0, addrSize, func(addr std.Address) bool {
				result += addr.String() + ", "
				return false
			})
			result = result[:len(result)-2]
		default:
			result += action + ufmt.Sprintf(": %d addresses", addrSize)
		}
		count++
		return false
	})
	return result
}
*/
