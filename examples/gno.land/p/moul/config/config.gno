package config

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/addrset"
	"gno.land/p/moul/authz"
)

var (
	// manager controls who can modify the list of keys.
	manager *authz.Authorizer
	// keys stores the list of managed addresses (keys).
	keys addrset.Set
)

func init() {
	// Initialize the manager with the deployer (origin) as the initial authority.
	manager = authz.NewWithOrigin()
	// Initialize the set to store the managed keys.
	keys = addrset.New()
}

// AddKey adds a new key to the managed list.
// Only authorized managers (defined by the authz.Authorizer) can call this.
func AddKey(key std.Address) error {
	if !key.IsValid() {
		return errors.New("invalid key address provided")
	}
	return manager.Do("add_key", func() error {
		keys.Add(key)
		// std.Emit("key_added", key.String()) // Optional: Emit an event
		return nil
	})
}

// RemoveKey removes a key from the managed list.
// Only authorized managers (defined by the authz.Authorizer) can call this.
func RemoveKey(key std.Address) error {
	if !key.IsValid() {
		return errors.New("invalid key address provided")
	}
	return manager.Do("remove_key", func() error {
		if !keys.Has(key) {
			return errors.New("key not found in the list")
		}
		keys.Remove(key)
		// std.Emit("key_removed", key.String()) // Optional: Emit an event
		return nil
	})
}

// HasKey checks if a key exists in the managed list.
// This is a public read-only function.
func HasKey(key std.Address) bool {
	return keys.Has(key)
}

// ListKeys returns a slice of all managed keys.
// This is a public read-only function.
func ListKeys() []std.Address {
	var keyList []std.Address
	// We need to cast to the underlying avl.Tree to iterate
	tree, ok := keys.Tree().(*avl.Tree)
	if !ok || tree == nil {
		return keyList // Return empty list if tree is not as expected or nil
	}
	tree.Iterate("", "", func(key string, _ any) bool {
		keyList = append(keyList, std.Address(key))
		return false // Continue iteration
	})
	return keyList
}

// --- Authority Management ---

// AddManager adds a new address to the list of authorized managers.
// This only works if the current authority is a MemberAuthority.
// The caller must be authorized by the current authority.
func AddManager(addr std.Address) error {
	memberAuth, ok := manager.Current().(*authz.MemberAuthority)
	if !ok {
		return errors.New("current authority is not a MemberAuthority, cannot add manager directly")
	}
	// Use the MemberAuthority's specific AddMember method,
	// which internally performs the authorization check.
	return memberAuth.AddMember(addr)
}

// RemoveManager removes an address from the list of authorized managers.
// This only works if the current authority is a MemberAuthority.
// The caller must be authorized by the current authority.
func RemoveManager(addr std.Address) error {
	memberAuth, ok := manager.Current().(*authz.MemberAuthority)
	if !ok {
		return errors.New("current authority is not a MemberAuthority, cannot remove manager directly")
	}
	// Use the MemberAuthority's specific RemoveMember method,
	// which internally performs the authorization check.
	return memberAuth.RemoveMember(addr)
}

// TransferManagement transfers the authority to manage keys to a new authority.
// The caller must be authorized by the current authority.
func TransferManagement(newAuthority authz.Authority) error {
	if newAuthority == nil {
		return errors.New("new authority cannot be nil")
	}
	// Use the Authorizer's Transfer method, which handles the authorization check.
	return manager.Transfer(newAuthority)
}

// GetManager returns a string representation of the current managing authority.
// This is a public read-only function.
func GetManager() string {
	return manager.String()
}
