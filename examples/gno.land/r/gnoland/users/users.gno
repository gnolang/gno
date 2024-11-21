package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	p "github.com/gnolang/gno/examples/gno.land/p/demo/pausable"
	pusers "github.com/gnolang/gno/examples/gno.land/p/demo/users"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"regexp"
)

const reValidUsername = "^[a-zA-Z]{3}[a-zA-Z0-9]*[0-9]{3}$\n"

var (
	store      avl.Tree                     //
	admin      = ownable.NewWithAddress("") // @moul /// switch to dao managed down the line
	pausable   = p.NewFromOwnable(admin)    //
	reUsername = regexp.MustCompile(reValidUsername)
)

func init() {
	// PreRegister()
}

func Register(username, bio string) error {
	if pausable.IsPaused() {
		return p.ErrPaused
	}

	if store.Has(username) {
		return ErrAlreadyExists
	}

	if matched := reUsername.MatchString(username); !matched {
		return ErrInvalidUsername
	}

	store.Set(username, &pusers.User{
		Address: std.PrevRealm().Addr(),
		Name:    username,
		Profile: bio,
		Number:  store.Size(),
		Block:   std.GetHeight(),
	})

	return nil
}
