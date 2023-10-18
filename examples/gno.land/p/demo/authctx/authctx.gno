// Package authctx offers a versatile framework for handling authentication contexts.
// It empowers the management of authentication contexts for users, realms, delegated calls,
// and potentially extends its capabilities to non-address entities and more.
package authctx

import (
	"std"

	"gno.land/p/demo/avl"
)

// AuthContext is an interface that represents an authentication context.
type AuthContext interface {
	Addr() std.Address
	// String returns a unique and human-readable identifier, typically in
	// the format "<type>:<addr>[/opts]".
	String() string
	HasRole(role string) bool
}

// AuthFunc signature should always match AuthContext.HasRole.
type AuthFunc func(role string) bool // TODO: switch to PBAC?
func AuthorizeAll() AuthFunc         { return func(_ string) bool { return true } }
func AuthorizeNone() AuthFunc        { return func(_ string) bool { return false } }
func AuthorizeRoles(roles ...string) AuthFunc {
	whitelist := map[string]bool{}
	for _, role := range roles {
		whitelist[role] = true
	}
	return func(check string) bool {
		return whitelist[check]
	}
}

// PermWrapper wraps an existing AuthContext to reduce its permissions while
// retaining previous ones.
type scopedAuthContext struct {
	parent      AuthContext
	authOverlay AuthFunc
}

func (s scopedAuthContext) Addr() std.Address { return s.parent.Addr() }
func (s scopedAuthContext) String() string    { return s.parent.String() } // XXX: +"/scoped"?
func (s scopedAuthContext) HasRole(role string) bool {
	return s.authOverlay(role) && s.parent.HasRole(role)
}

func NewScopedAuthContext(ctx AuthContext, authOverlay AuthFunc) AuthContext {
	return &scopedAuthContext{
		parent:      ctx,
		authOverlay: authOverlay,
	}
}

// TODO: Add a helper to create an exposed helper allowing remote contracts to compose their authentication context.

// origAuthContext represents an authentication context based on `std.GetOrigCaller`.
// It represents the address of the account used to create the transrole.
// This is true even if there are multiple packages and realms called in between.
type origAuthContext struct {
	addr std.Address
}

func NewOrigAuthContext() AuthContext {
	return origAuthContext{addr: std.GetOrigCaller()}
}
func (a origAuthContext) Addr() std.Address        { return a.addr }
func (a origAuthContext) String() string           { return "orig:" + string(a.Addr()) }
func (a origAuthContext) HasRole(role string) bool { return true }

// prevAuthContext represents an authentication context based on `std.PrevRealm()`.
// It represents the address of the previous realm in the stack.
// This can be another realm or the caller (same as OrigAuthContext) in the case
// of a direct call to the helper without an intermediary realm.
type prevAuthContext struct {
	addr std.Address
}

func NewPrevAuthContext() AuthContext {
	return prevAuthContext{std.PrevRealm().Addr()}
}
func (a prevAuthContext) Addr() std.Address        { return a.addr }
func (a prevAuthContext) String() string           { return "prev:" + string(a.Addr()) }
func (a prevAuthContext) HasRole(role string) bool { return true }

// delegatedAuthContext represents an authentication context that allows someone to approve other addresses to post on its behalf.
type delegatedAuthContext struct {
	owner  std.Address
	caller std.Address
	authFn AuthFunc
}

func (a delegatedAuthContext) Addr() std.Address { return a.owner }
func (a delegatedAuthContext) String() string {
	return "delegated:" + string(a.Addr()) + "/" + string(a.caller)
}
func (a delegatedAuthContext) HasRole(role string) bool { return a.authFn(role) }

// DelegationDB is a type that represents a database of delegations.
// It maps owner addresses to their respective delegates.
// This structure is intended to be initialized and stored globally by a realm
// that wants to offer delegated authentication context to its users.
type DelegationDB struct{ tree avl.Tree } // owner(std.Address) -> delegates(avl.Tree)

// delegates is a type that represents a group of delegates.
// It maps delegate addresses to their respective isAuthorizedFn.
type (
	delegates struct{ tree avl.Tree } // delegate(std.Address) -> AuthFunc
	delegate  struct{ authFn AuthFunc }
)

func (d *DelegationDB) Approve(delegateAddr std.Address, authFn AuthFunc) {
	// std.AssertOriginCall() // TODO: re-enable when compatible with `go test`
	owner := std.GetOrigCaller()
	delegations, _ := d.delegationsByOwner(owner)
	delegations.tree.Set(string(delegateAddr), delegate{authFn})
	d.tree.Set(string(owner), delegations)
}

func (d DelegationDB) delegationsByOwner(owner std.Address) (delegates, bool) {
	delegations, found := d.tree.Get(string(owner))
	if found {
		return delegations.(delegates), true
	}
	return delegates{}, false
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
	delegateRaw, found := delegations.tree.Get(string(caller))
	if !found {
		return nil
	}
	authFn := delegateRaw.(delegate).authFn

	return &delegatedAuthContext{
		owner:  owner,
		caller: caller,
		authFn: authFn,
	}
}

func Must(ctx AuthContext) AuthContext {
	if ctx == nil {
		panic("unauthorized")
	}
	return ctx
}
