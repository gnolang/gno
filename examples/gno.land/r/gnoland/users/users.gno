package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	p "github.com/gnolang/gno/examples/gno.land/p/demo/pausable"
	pusers "github.com/gnolang/gno/examples/gno.land/p/demo/users"
	sys "github.com/gnolang/gno/examples/gno.land/r/sys/users"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"regexp"
)

const (
	reValidUsername = "^[a-zA-Z]{3}[a-zA-Z0-9]{0,14}[0-9]{3}$"
	RegisterEvent   = "Register"
	adminAddr       = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul
)

var (
	admin      = ownable.NewWithAddress(adminAddr) // @moul /// switch to dao managed down the line
	pausable   = p.NewFromOwnable(admin)           //
	reUsername = regexp.MustCompile(reValidUsername)
)

// Register registers a new username
// A valid username must start with a minimum of 3 letters,
// end with a minimum of 3 numbers, and be less than 20 chars long.
func Register(username, bio string) error {
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
