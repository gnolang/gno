// Package authctx provides a unified way of representing authentication context.
// It supports various workflows and allows the calling package to configure AuthContext as needed.
// For instance, some realms might disable OrigCaller or only keep DelegatedAuthContext.
package authctx

import (
	"std"

	"gno.land/p/demo/avl"
)

// AuthContext is an interface that represents an authentication context.
type AuthContext interface {
	Addr() std.Address
	String() string
	HasPerm(perm string) bool
}

// TODO: Add a helper to create an exposed helper allowing remote contracts to compose their authentication context.

// permKey is a type for permission keys.
type permKey string

const PermAll = permKey("all")

// origAuthContext represents an authentication context based on `std.GetOrigCaller`.
// It represents the address of the account used to create the transaction.
// This is true even if there are multiple packages and realms called in between.
type origAuthContext struct{ addr std.Address }

func NewOrigAuthContext() AuthContext           { return origAuthContext{std.GetOrigCaller()} }
func (a origAuthContext) Addr() std.Address     { return a.addr }
func (a origAuthContext) String() string        { return "orig:" + string(a.Addr()) }
func (a origAuthContext) HasPerm(_ string) bool { return true }

// prevAuthContext represents an authentication context based on `std.PrevRealm()`.
// It represents the address of the previous realm in the stack.
// This can be another realm or the caller (same as OrigAuthContext) in the case
// of a direct call to the helper without an intermediary realm.
type prevAuthContext struct{ addr std.Address }

func NewPrevAuthContext() AuthContext           { return prevAuthContext{std.PrevRealm().Addr()} }
func (a prevAuthContext) Addr() std.Address     { return a.addr }
func (a prevAuthContext) String() string        { return "prev:" + string(a.Addr()) }
func (a prevAuthContext) HasPerm(_ string) bool { return true }

// delegatedAuthContext represents an authentication context that allows someone to approve other addresses to post on its behalf.
type delegatedAuthContext struct {
	owner  std.Address
	caller std.Address
	perms  []string
}

func (a delegatedAuthContext) Addr() std.Address { return a.owner }
func (a delegatedAuthContext) String() string {
	return "delegated:" + string(a.Addr()) + "(" + string(a.caller) + ")"
}

func (a delegatedAuthContext) HasPerm(perm string) bool {
	for _, p := range a.perms {
		if p == string(PermAll) || p == perm {
			return true
		}
	}
	return false
}

// DelegationDB is a type that represents a database of delegations.
// It maps owner addresses to their respective delegates.
// This structure is intended to be initialized and stored globally by a realm
// that wants to offer delegated authentication context to its users.
type DelegationDB struct{ tree avl.Tree } // owner(std.Address) -> delegates(avl.Tree)

// delegates is a type that represents a group of delegates.
// It maps delegate addresses to their respective permissions.
type delegates struct{ tree avl.Tree } // delegate(std.Address) -> perms([]Perm)

func (d *DelegationDB) Approve(delegateAddr std.Address, newPerms ...string) {
	std.AssertOriginCall()
	owner := std.GetOrigCaller()
	delegations, _ := d.delegationsByOwner(owner)
	perms, _ := delegations.permsByAddr(delegateAddr)
	for _, newPerm := range newPerms {
		// TODO: Check for duplicate perms.
		perms = append(perms, newPerm)
	}
	delegations.tree.Set(string(delegateAddr), perms)
	d.tree.Set(string(owner), delegations)
}

func (d DelegationDB) delegationsByOwner(owner std.Address) (delegates, bool) {
	delegations, found := d.tree.Get(string(owner))
	if found {
		return delegations.(delegates), true
	}
	return delegates{}, false
}

func (d delegates) permsByAddr(addr std.Address) ([]string, bool) {
	perms, found := d.tree.Get(string(addr))
	if found {
		return perms.([]string), true
	}
	return []string{}, false
}

// NewOrigDelegatedAuthContext creates a new delegatedAuthContext object.
// It is intended to be called by the pre-approved delegate.
// It returns a delegatedAuthContext object if it was approved; else it panics.
func (d DelegationDB) NewOrigDelegatedAuthContext(owner std.Address) AuthContext {
	caller := std.GetOrigCaller()
	return d.newDelegatedAuthContext(caller, owner)
}

func (d DelegationDB) NewPrevDelegatedAuthContext(owner std.Address) AuthContext {
	caller := std.PrevRealm().Addr()
	return d.newDelegatedAuthContext(caller, owner)
}

func (d DelegationDB) newDelegatedAuthContext(caller, owner std.Address) AuthContext {
	delegations, found := d.delegationsByOwner(owner)
	if !found {
		return nil
	}
	perms, found := delegations.permsByAddr(caller)
	if !found {
		return nil
	}

	return &delegatedAuthContext{
		owner:  owner,
		caller: caller,
		perms:  perms,
	}
}

func Must(auth AuthContext) AuthContext {
	if auth == nil {
		panic("unauthorized")
	}
	return auth
}
