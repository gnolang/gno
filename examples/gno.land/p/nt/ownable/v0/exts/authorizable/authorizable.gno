// Package authorizable is an extension of p/nt/ownable;
// It allows the user to instantiate an Authorizable struct, which extends
// p/nt/ownable with a list of users that are authorized for something.
// By using authorizable, you have a superuser (ownable), as well as another
// authorization level, which can be used for adding moderators or similar to your realm.
package authorizable

import (
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/ufmt/v0"
)

type Authorizable struct {
	*ownable.Ownable                // owner in ownable is superuser
	authorized       *bptree.BPTree // chain.Addr > struct{}{}
}

// New creates an Authorizable from an existing *ownable.Ownable.
// The owner is automatically added to the auth list.
//
// Example construction:
//
//	authorizable.New(ownable.NewWithAddress(addr))
func New(o *ownable.Ownable) *Authorizable {
	a := &Authorizable{
		Ownable:    o,
		authorized: bptree.NewBPTree32(),
	}

	// Add owner to auth list
	a.authorized.Set(a.Owner().String(), struct{}{})
	return a
}

// AddToAuthList adds addr to the auth list. rlm must be the caller's
// own captured cur; rlm.Previous().Address() must equal the superuser
// (the underlying Ownable's owner).
func (a *Authorizable) AddToAuthList(_ int, rlm realm, addr address) error {
	if !rlm.IsCurrent() {
		return ErrNotSuperuser
	}
	if !a.OwnedBy(rlm.Previous().Address()) {
		return ErrNotSuperuser
	}
	return a.addToAuthList(addr)
}

func (a *Authorizable) addToAuthList(addr address) error {
	if _, exists := a.authorized.Get(addr.String()); exists {
		return ErrAlreadyInList
	}

	a.authorized.Set(addr.String(), struct{}{})

	return nil
}

// DeleteFromAuthList removes addr from the auth list. rlm must be the
// caller's own captured cur; rlm.Previous().Address() must equal the
// superuser (the underlying Ownable's owner).
func (a *Authorizable) DeleteFromAuthList(_ int, rlm realm, addr address) error {
	if !rlm.IsCurrent() {
		return ErrNotSuperuser
	}
	if !a.OwnedBy(rlm.Previous().Address()) {
		return ErrNotSuperuser
	}
	return a.deleteFromAuthList(addr)
}

func (a *Authorizable) deleteFromAuthList(addr address) error {
	if !a.authorized.Has(addr.String()) {
		return ErrNotInAuthList
	}

	if _, removed := a.authorized.Remove(addr.String()); !removed {
		str := ufmt.Sprintf("authorizable: could not remove %s from auth list", addr.String())
		panic(str)
	}

	return nil
}

// OnAuthList reports whether rlm.Address() is on the auth list. rlm
// must be the caller's own captured cur (asserted via rlm.IsCurrent()).
// Pre-migration shape used runtime.CurrentRealm().Address() — vulnerable
// to the .Title()-class read where a non-crossing wrapper made the walk
// return the wrong realm. Explicit rlm closes that.
func (a *Authorizable) OnAuthList(_ int, rlm realm) error {
	if !rlm.IsCurrent() {
		return ErrNotInAuthList
	}
	return a.onAuthList(rlm.Address())
}

// PreviousOnAuthList reports whether rlm.Previous().Address() — the
// realm that crossed into the caller — is on the auth list. Same rlm
// contract as OnAuthList.
func (a *Authorizable) PreviousOnAuthList(_ int, rlm realm) error {
	if !rlm.IsCurrent() {
		return ErrNotInAuthList
	}
	return a.onAuthList(rlm.Previous().Address())
}

func (a *Authorizable) onAuthList(caller address) error {
	if !a.authorized.Has(caller.String()) {
		return ErrNotInAuthList
	}
	return nil
}

func (a Authorizable) AssertOnAuthList(_ int, rlm realm) {
	if err := a.OnAuthList(0, rlm); err != nil {
		panic(err)
	}
}

func (a Authorizable) AssertPreviousOnAuthList(_ int, rlm realm) {
	if err := a.PreviousOnAuthList(0, rlm); err != nil {
		panic(err)
	}
}
