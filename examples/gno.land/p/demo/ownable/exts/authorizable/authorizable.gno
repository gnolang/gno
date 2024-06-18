// Package authorizable is an extension of p/demo/ownable;
// It allows the user to instantiate an Authorizable struct, which extends
// p/demo/ownable with a list of users that are authorized for something.
// By using authorizable, you have a superuser (ownable), as well as another
// authorization level, which can be used for adding moderators or similar to your realm.
package authorizable

import (
	"std"

	"gno.land/p/demo/ownable"
)

type Authorizable struct {
	*ownable.Ownable // owner in ownable is superuser
	authorized       map[std.Address]struct{}
}

func NewAuthorizable() *Authorizable {
	a := &Authorizable{
		ownable.NewOwnable(),
		make(map[std.Address]struct{}),
	}

	// Add owner to auth list
	a.authorized[a.Owner()] = struct{}{}
	return a
}

func NewAuthorizableWithAddress(addr std.Address) *Authorizable {
	a := &Authorizable{
		ownable.NewOwnableWithAddress(addr),
		make(map[std.Address]struct{}),
	}

	// Add owner to auth list
	a.authorized[addr] = struct{}{}
	return a
}

func (a *Authorizable) AddToAuthList(addr std.Address) error {
	if err := a.CallerIsOwner(); err != nil {
		return errNotSuperuser
	}

	if _, exists := a.authorized[addr]; exists {
		return errAlreadyInList
	}

	a.authorized[addr] = struct{}{}
	return nil
}

func (a Authorizable) CallerOnAuthList() error {
	caller := std.PrevRealm().Addr()
	if _, exists := a.authorized[caller]; exists {
		return nil
	}

	return errNotInAuthList
}

func (a Authorizable) OnAuthListByAddr(addr std.Address) error {
	caller := std.PrevRealm().Addr()
	if _, exists := a.authorized[caller]; exists {
		return nil
	}

	return errNotInAuthList
}

func (a Authorizable) AssertOnAuthList() {
	caller := std.PrevRealm().Addr()

	if _, exists := a.authorized[caller]; !exists {
		panic(errNotInAuthList)
	}
}

// todo add remove from auth list
