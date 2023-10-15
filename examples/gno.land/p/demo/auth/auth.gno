// Package auth provides a unified way of representing authentication mechanisms.
// It supports various workflows and allows the calling package to configure Auth as needed.
// For instance, some realms might disable OrigCaller or only keep DelegatedAuth.
package auth

import (
	"std"

	"gno.land/p/demo/avl"
)

// Auth is an interface that represents an authentication method.
type Auth interface {
	Addr() std.Address
	Kind() string
	String() string
	HasPerm(perm string) bool
}

// TODO: add a helper to create an exposed helper allowing remote contracts to compose their authentication.

// permKey is a type for permission keys.
type permKey string

const PermAll = permKey("all")

// origAuth is an authentication method based on `std.GetOrigCaller`.
// It represents the address of the account used to create the transaction.
// This is true even if there are multiple packages and realms called in between.
type origAuth struct{ addr std.Address }

func NewOrigAuth() Auth                  { return origAuth{std.GetOrigCaller()} }
func (a origAuth) Addr() std.Address     { return a.addr }
func (a origAuth) Kind() string          { return "orig" }
func (a origAuth) String() string        { return a.Kind() + ":" + string(a.Addr()) }
func (a origAuth) HasPerm(_ string) bool { return true }

// prevAuth is an authentication method based on `std.PrevRealm()`.
// It represents the address of the previous realm in the stack.
// This can be another realm or the caller (same as OrigAuth) in the case
// of a direct call to the helper without an intermediary realm.
type prevAuth struct{ addr std.Address }

func NewPrevAuth() Auth                  { return prevAuth{std.PrevRealm().Addr()} }
func (a prevAuth) Addr() std.Address     { return a.addr }
func (a prevAuth) Kind() string          { return "prev" }
func (a prevAuth) String() string        { return a.Kind() + ":" + string(a.Addr()) }
func (a prevAuth) HasPerm(_ string) bool { return true }

// delegatedAuth is an authentication method that allows someone to approve other addresses to post on its behalf.
type delegatedAuth struct {
	owner  std.Address
	caller std.Address
	perms  []string
}

func (a delegatedAuth) Addr() std.Address { return a.owner }
func (a delegatedAuth) Kind() string      { return "delegated" }
func (a delegatedAuth) String() string {
	return a.Kind() + ":" + string(a.Addr()) + "(" + string(a.caller) + ")"
}

func (a delegatedAuth) HasPerm(perm string) bool {
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
// that wants to offer delegated authentication to its users.
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
		// TODO: check for duplicate perms
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

// NewOrigDelegatedAuth creates a new delegatedAuth object.
// It is intended to be called by the pre-approved delegate.
// It returns a delegatedAuth object if it was approved; else it panics.
func (d DelegationDB) NewOrigDelegatedAuth(owner std.Address) Auth {
	caller := std.GetOrigCaller()
	return d.newDelegatedAuth(caller, owner)
}

func (d DelegationDB) NewPrevDelegatedAuth(owner std.Address) Auth {
	caller := std.PrevRealm().Addr()
	return d.newDelegatedAuth(caller, owner)
}

func (d DelegationDB) newDelegatedAuth(caller, owner std.Address) Auth {
	delegations, found := d.delegationsByOwner(owner)
	if !found {
		return nil
	}
	perms, found := delegations.permsByAddr(caller)
	if !found {
		return nil
	}

	return &delegatedAuth{
		owner:  owner,
		caller: caller,
		perms:  perms,
	}
}

func Must(auth Auth) Auth {
	if auth == nil {
		panic("unauthorized")
	}
	return auth
}
