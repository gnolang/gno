package config

import (
	"errors"
	"std"

	"gno.land/p/moul/authz"
)

var Authorizer = authz.NewWithOrigin()

// AddManager adds a new address to the list of authorized managers.
// This only works if the current authority is a MemberAuthority.
// The caller must be authorized by the current authority.
func AddManager(addr std.Address) error {
	crossing()
	caller := std.PreviousRealm().Address()
	memberAuth, ok := Authorizer.Authority().(*authz.MemberAuthority)
	if !ok {
		return errors.New("current authority is not a MemberAuthority, cannot add manager directly")
	}
	// Use the MemberAuthority's specific AddMember method,
	// which internally performs the authorization check.
	return memberAuth.AddMember(caller, addr)
}

// RemoveManager removes an address from the list of authorized managers.
// This only works if the current authority is a MemberAuthority.
// The caller must be authorized by the current authority.
func RemoveManager(addr std.Address) error {
	crossing()
	caller := std.PreviousRealm().Address()
	memberAuth, ok := Authorizer.Authority().(*authz.MemberAuthority)
	if !ok {
		return errors.New("current authority is not a MemberAuthority, cannot remove manager directly")
	}
	// Use the MemberAuthority's specific RemoveMember method,
	// which internally performs the authorization check.
	return memberAuth.RemoveMember(caller, addr)
}

// TransferManagement transfers the authority to manage keys to a new authority.
// The caller must be authorized by the current authority.
func TransferManagement(newAuthority authz.Authority) error {
	crossing()
	caller := std.PreviousRealm().Address()
	if newAuthority == nil {
		return errors.New("new authority cannot be nil")
	}
	// Use the Authorizer's Transfer method, which handles the authorization check.
	return Authorizer.Transfer(caller, newAuthority)
}

// ListManagers returns a slice of all managed keys.
func ListManagers() []std.Address {
	crossing()
	var keyList []std.Address

	memberAuth, ok := Authorizer.Authority().(*authz.MemberAuthority)
	if !ok {
		return keyList
	}
	tree := memberAuth.Tree()
	if !ok || tree == nil {
		return keyList // Return empty list if tree is not as expected or nil
	}
	tree.Iterate("", "", func(key string, _ any) bool {
		keyList = append(keyList, std.Address(key))
		return false
	})
	return keyList
}

func HasManager(addr std.Address) bool {
	crossing()
	memberAuth, ok := Authorizer.Authority().(*authz.MemberAuthority)
	if !ok {
		return false // Return false if not a MemberAuthority or doesn't exist
	}
	// Use the MemberAuthority's specific RemoveMember method,
	// which internally performs the authorization check.
	return memberAuth.Has(addr)
}
