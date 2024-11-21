package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	p "github.com/gnolang/gno/examples/gno.land/p/demo/pausable"
	pusers "github.com/gnolang/gno/examples/gno.land/p/demo/users"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

var (
	uStore          = avl.NewTree() // username > *pusers.User
	aStore          = avl.NewTree() // address > *pusers.User
	callerWhitelist = avl.NewTree() // addr > struct{}{}
)

func WriteUser(user *pusers.User) error {
	if pausable.IsPaused() {
		return p.ErrPaused
	}

	if !callerWhitelist.Has(std.PrevRealm().Addr().String()) {
		return ErrUnauthorized
	}

	username := user.Name
	addr := user.Address

	if uStore.Has(username) {
		return ErrAlreadyExists
	}

	uStore.Set(username, user)
	aStore.Set(addr.String(), user)

	return nil
}

func UserCount() int {
	return uStore.Size()
}
