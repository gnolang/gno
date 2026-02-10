package config

import (
	"chain/runtime"
	"errors"

	"gno.land/p/moul/authz"
)

var Authorizer = authz.NewWithOrigin()

// AddManager adds a new address to the list of authorized managers.
// This only works if the current authority is a MemberAuthority.
// The caller must be authorized by the current authority.
func AddManager(cur realm, addr address) error {
	caller := runtime.PreviousRealm().Address()
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
func RemoveManager(cur realm, addr address) error {
	caller := runtime.PreviousRealm().Address()
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
func TransferManagement(cur realm, newAuthority authz.Authority) error {
	caller := runtime.PreviousRealm().Address()
	if newAuthority == nil {
		return errors.New("new authority cannot be nil")
	}
	// Use the Authorizer's Transfer method, which handles the authorization check.
	return Authorizer.Transfer(caller, newAuthority)
}

// ListManagers returns a slice of all managed keys.
func ListManagers(cur realm) []address {
	var keyList []address
	memberAuth, ok := Authorizer.Authority().(*authz.MemberAuthority)
	if !ok {
		return keyList
	}
	tree := memberAuth.Tree()
	if !ok || tree == nil {
		return keyList // Return empty list if tree is not as expected or nil
	}
	tree.Iterate("", "", func(key string, _ any) bool {
		keyList = append(keyList, address(key))
		return false
	})
	return keyList
}

func HasManager(cur realm, addr address) bool {
	memberAuth, ok := Authorizer.Authority().(*authz.MemberAuthority)
	if !ok {
		return false // Return false if not a MemberAuthority or doesn't exist
	}
	// Use the MemberAuthority's specific RemoveMember method,
	// which internally performs the authorization check.
	return memberAuth.Has(addr)
}
