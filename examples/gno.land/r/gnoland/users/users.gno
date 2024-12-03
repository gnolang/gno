package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	p "gno.land/p/demo/pausable"
	susers "gno.land/p/sys/users"
	"regexp"
	"std"

	sys "gno.land/r/sys/users"
)

const (
	reValidUsername = "^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}$"
	adminAddr       = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul
	RegisterEvent   = "Register"
)

var (
	admin             = ownable.NewWithAddress(adminAddr) // @moul /// switch to dao managed down the line
	pausable          = p.NewFromOwnable(admin)           //
	reUsername        = regexp.MustCompile(reValidUsername)
	extendedUserStore = avl.NewTree() // addr > *ExtendedUser
)

// Register registers a new username for the caller
// A valid username must start with a minimum of 3 letters,
// end with a minimum of 3 numbers, and be less than 20 chars long.
// All letters must be lowercase, and the only valid special char is `_`
// Only calls from EOAs are supported
func Register(username, bio string) error {
	std.AssertOriginCall()

	if pausable.IsPaused() {
		return p.ErrPaused
	}

	if matched := reUsername.MatchString(username); !matched {
		return ErrInvalidUsername
	}

	registrant := std.PrevRealm().Addr()

	base := &susers.User{
		Address: registrant,
		Name:    username,
	}

	extendedUser := &User{
		base:         base,
		Bio:          bio, // optional
		Ordinal:      sys.UserCount(),
		RegisteredAt: std.GetHeight(),
	}

	err := sys.WriteUser(base)
	if err != nil {
		return err
	}

	extendedUserStore.Set(registrant.String(), extendedUser)

	std.Emit(RegisterEvent, "registrant", registrant.String(), "username", username)
	return nil
}

func UpdateBio(bio string) error {
	if pausable.IsPaused() {
		return p.ErrPaused
	}

	raw, ok := extendedUserStore.Get(std.PrevRealm.Addr())
	if !ok {
		return ErrNotRegistered
	}

	user := raw.(*User)
	user.Bio = bio

	return nil
}
