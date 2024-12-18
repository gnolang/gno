package users

import (
	"regexp"
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/pausable"

	"gno.land/r/sys/users"
)

const (
	reValidUsername = "^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}$"
	adminAddr       = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul
	RegisterEvent   = "Register"
)

var (
	Ownable    = ownable.NewWithAddress(adminAddr) // @moul // switch to dao managed down the line
	Pausable   = pausable.NewFromOwnable(Ownable)  //
	reUsername = regexp.MustCompile(reValidUsername)
)

// Register registers a new username for the caller
// A valid username must start with a minimum of 3 letters,
// end with a minimum of 3 numbers, and be less than 20 chars long.
// All letters must be lowercase, and the only valid special char is `_`
// Only calls from EOAs are supported
func Register(username string) error {
	std.AssertOriginCall()

	if Pausable.IsPaused() {
		return p.ErrPaused
	}

	if matched := reUsername.MatchString(username); !matched {
		return ErrInvalidUsername
	}

	registrant := std.PrevRealm().Addr()

	err := users.WriteUser(username, registrant)
	if err != nil {
		return err
	}

	std.Emit(RegisterEvent, "registrant", registrant.String(), "username", username)
	return nil
}
