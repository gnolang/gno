package users

import (
	"regexp"
	"std"

	susers "gno.land/r/sys/users"
)

const (
	reValidUsername = "^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}$"
)

var reUsername = regexp.MustCompile(reValidUsername)

// Register registers a new username for the caller
// A valid username must start with a minimum of 3 letters,
// end with a minimum of 3 numbers, and be less than 20 chars long.
// All letters must be lowercase, and the only valid special char is `_`.
// Only calls from EOAs are supported.
func Register(username string) error {
	std.AssertOriginCall()

	if paused {
		return ErrPaused
	}

	if matched := reUsername.MatchString(username); !matched {
		return ErrInvalidUsername
	}

	registrant := std.PreviousRealm().Address()
	if err := susers.RegisterUser(username, registrant); err != nil {
		return err
	}

	return nil
}
