// Package authorizable is an extension of p/demo/ownable;
// It allows the user to instantiate an Authorizable struct, which extends
// p/demo/ownable with a list of users that are authorized for something.
// By using authorizable, you have a superuser (ownable), as well as another
// authorization level, which can be used for adding moderators or similar to your realm.
package authorizable

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

type Authorizable struct {
	*ownable.Ownable           // owner in ownable is superuser
	authorized       *avl.Tree // std.Addr > struct{}{}
}

func NewAuthorizable() *Authorizable {
	a := &Authorizable{
		ownable.New(),
		avl.NewTree(),
	}

	// Add owner to auth list
	a.authorized.Set(a.Owner().String(), struct{}{})
	return a
}

func NewAuthorizableWithAddress(addr std.Address) *Authorizable {
	a := &Authorizable{
		ownable.NewWithAddress(addr),
		avl.NewTree(),
	}

	// Add owner to auth list
	a.authorized.Set(a.Owner().String(), struct{}{})
	return a
}

func (a *Authorizable) AddToAuthList(addr std.Address) error {
	if err := a.CallerIsOwner(); err != nil {
		return ErrNotSuperuser
	}

	if _, exists := a.authorized.Get(addr.String()); exists {
		return ErrAlreadyInList
	}

	a.authorized.Set(addr.String(), struct{}{})

	return nil
}

func (a *Authorizable) DeleteFromAuthList(addr std.Address) error {
	if err := a.CallerIsOwner(); err != nil {
		return ErrNotSuperuser
	}

	if !a.authorized.Has(addr.String()) {
		return ErrNotInAuthList
	}

	if _, removed := a.authorized.Remove(addr.String()); !removed {
		str := ufmt.Sprintf("authorizable: could not remove %s from auth list", addr.String())
		panic(str)
	}

	return nil
}

func (a Authorizable) CallerOnAuthList() error {
	caller := std.PrevRealm().Addr()

	if !a.authorized.Has(caller.String()) {
		return ErrNotInAuthList
	}

	return nil
}

func (a Authorizable) AssertOnAuthList() {
	caller := std.PrevRealm().Addr()

	if !a.authorized.Has(caller.String()) {
		panic(ErrNotInAuthList)
	}
}
