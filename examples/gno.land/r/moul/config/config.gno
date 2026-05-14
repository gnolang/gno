package config

import (
	"errors"

	"gno.land/p/moul/authz"
)

var Authorizer *authz.Authorizer

func init(cur realm) {
	if !cur.Previous().IsUserCall() {
		panic("r/moul/config must be initialized by an EOA")
	}
	Authorizer = authz.NewWithMembers(cur.Previous().Address())
}

// AddManager adds a new address to the list of authorized managers.
// This only works if the current authority is a MemberAuthority.
// The caller must be authorized by the current authority.
func AddManager(cur realm, addr address) error {
	memberAuth, ok := Authorizer.Authority().(*authz.MemberAuthority)
	if !ok {
		return errors.New("current authority is not a MemberAuthority, cannot add manager directly")
	}
	return memberAuth.AddMember(0, cur, addr)
}

// RemoveManager removes an address from the list of authorized managers.
// This only works if the current authority is a MemberAuthority.
// The caller must be authorized by the current authority.
func RemoveManager(cur realm, addr address) error {
	memberAuth, ok := Authorizer.Authority().(*authz.MemberAuthority)
	if !ok {
		return errors.New("current authority is not a MemberAuthority, cannot remove manager directly")
	}
	return memberAuth.RemoveMember(0, cur, addr)
}

// SECURITY XXX: TransferManagement accepts any Authority impl as the
// new authority — `newAuthority` is an open interface. Two risks:
// (a) If the caller passes a foreign Authority impl, the impl's
//     Authorize method runs with whatever this realm's authority
//     grants. Hostile impls can always-approve (silent privilege
//     escalation) or always-deny (DoS / lockout).
// (b) Once the new authority is installed, there is no way back if it
//     turns out hostile — the new authority controls all future
//     Transfer calls. An attacker who becomes the authority (legit
//     or via a malicious newAuthority) can install a permanent
//     DroppedAuthority to lock this realm forever.
// Fix: restrict newAuthority to canonical authz impls (MemberAuthority,
// ContractAuthority, etc. from p/moul/authz). Optionally wrap Transfer
// with a quorum/cooldown check or a one-shot guard so a single
// compromised authority can't permanently brick the config realm.
//
// TransferManagement transfers the authority to manage keys to a new authority.
// The caller must be authorized by the current authority.
func TransferManagement(cur realm, newAuthority authz.Authority) error {
	if newAuthority == nil {
		return errors.New("new authority cannot be nil")
	}
	return Authorizer.Transfer(0, cur, newAuthority)
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
