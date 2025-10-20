// Package authorizable is an extension of p/demo/ownable;
// It allows the user to instantiate an Authorizable struct, which extends
// p/demo/ownable with a list of users that are authorized for something.
// By using authorizable, you have a superuser (ownable), as well as another
// authorization level, which can be used for adding moderators or similar to your realm.
package authorizable

import (
	"chain/runtime"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/ufmt"
)

type Authorizable struct {
	*ownable.Ownable           // owner in ownable is superuser
	authorized       *avl.Tree // chain.Addr > struct{}{}
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

func NewAuthorizableWithOrigin() *Authorizable {
	a := &Authorizable{
		ownable.NewWithOrigin(),
		avl.NewTree(),
	}

	// Add owner to auth list
	a.authorized.Set(a.Owner().String(), struct{}{})
	return a
}

func NewAuthorizableWithAddress(addr address) *Authorizable {
	a := &Authorizable{
		ownable.NewWithAddress(addr),
		avl.NewTree(),
	}

	// Add owner to auth list
	a.authorized.Set(a.Owner().String(), struct{}{})
	return a
}

func (a *Authorizable) AddToAuthList(addr address) error {
	if !a.OwnedByCurrent() {
		return ErrNotSuperuser
	}
	return a.addToAuthList(addr)
}

func (a *Authorizable) AddToAuthListByPrevious(addr address) error {
	if !a.OwnedByPrevious() {
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

func (a *Authorizable) DeleteFromAuthList(addr address) error {
	if !a.OwnedByCurrent() {
		return ErrNotSuperuser
	}
	return a.deleteFromAuthList(addr)
}

func (a *Authorizable) DeleteFromAuthListByPrevious(addr address) error {
	if !a.OwnedByPrevious() {
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

func (a *Authorizable) OnAuthList() error {
	current := runtime.CurrentRealm().Address()
	return a.onAuthList(current)
}

func (a *Authorizable) PreviousOnAuthList() error {
	previous := runtime.PreviousRealm().Address()
	return a.onAuthList(previous)
}

func (a *Authorizable) onAuthList(caller address) error {
	if !a.authorized.Has(caller.String()) {
		return ErrNotInAuthList
	}
	return nil
}

func (a Authorizable) AssertOnAuthList() {
	err := a.OnAuthList()
	if err != nil {
		panic(err)
	}
}

func (a Authorizable) AssertPreviousOnAuthList() {
	err := a.PreviousOnAuthList()
	if err != nil {
		panic(err)
	}
}
