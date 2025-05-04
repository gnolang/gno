package users

import (
	"regexp"
	"std"

	"gno.land/p/moul/fifo"
	susers "gno.land/r/sys/users"
)

const (
	reValidUsername = "^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}$"
)

var (
	registerPrice = int64(1_000_000) // 1 GNOT
	latestUsers   = fifo.New(10)     // Save the latest 10 users for rendering purposes
	reUsername    = regexp.MustCompile(reValidUsername)
)

// Register registers a new username for the caller.
// A valid username must start with a minimum of 3 letters,
// end with a minimum of 3 numbers, and be less than 20 chars long.
// All letters must be lowercase, and the only valid special char is `_`.
// Only calls from EOAs are supported.
func Register(username string) {
	crossing()

	if !std.PreviousRealm().IsUser() {
		panic(ErrNonUserCall.Error())
	}

	if paused {
		panic(ErrPaused.Error())
	}

	if std.OriginSend().AmountOf("ugnot") != registerPrice {
		panic(ErrInvalidPayment.Error())
	}

	if matched := reUsername.MatchString(username); !matched {
		panic(ErrInvalidUsername.Error())
	}

	registrant := std.PreviousRealm().Address()
	if err := cross(susers.RegisterUser)(username, registrant); err != nil {
		panic(err.Error())
	}

	latestUsers.Append(username)
	std.Emit("Registeration", "address", registrant.String(), "name", username)
}
