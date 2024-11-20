package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	p "github.com/gnolang/gno/examples/gno.land/p/demo/pausable"
	"github.com/gnolang/gno/examples/gno.land/p/demo/seqid"
	"regexp"
)

const reValidUsername = "^[a-zA-Z]{3}[a-zA-Z0-9]*[0-9]{3}$\n"

var (
	store      avl.Tree //
	idCounter  seqid.ID
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

	matched, _ := regexp.MatchString(reValidUsername, username)
	if !matched {
		return ErrInvalidUsername
	}

	store.Set(username)

	return nil
}
