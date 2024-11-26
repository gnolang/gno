package users

import (
	"regexp"
	"std"

	"gno.land/p/demo/ownable"
	p "gno.land/p/demo/pausable"
	pusers "gno.land/p/demo/users"

	sys "gno.land/r/sys/users"
)

const (
	reValidUsername = "^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}$"
	adminAddr       = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul
	RegisterEvent   = "Register"
)

var (
	admin      = ownable.NewWithAddress(adminAddr) // @moul /// switch to dao managed down the line
	pausable   = p.NewFromOwnable(admin)           //
	reUsername = regexp.MustCompile(reValidUsername)
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
	u := &pusers.User{
		Address: registrant,
		Name:    username,
		Profile: bio, // optional
		Number:  sys.UserCount(),
		Block:   std.GetHeight(),
	}

	err := sys.WriteUser(u)
	if err != nil {
		return err
	}

	std.Emit(RegisterEvent, "registrant", registrant.String(), "username", username)
	return nil
}
