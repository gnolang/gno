package users

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"regexp"

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
func Register(_ realm, username string) {
	if !runtime.PreviousRealm().IsUser() {
		panic(ErrNonUserCall)
	}

	if paused {
		panic(ErrPaused)
	}

	if banker.OriginSend().AmountOf("ugnot") != registerPrice {
		panic(ErrInvalidPayment)
	}

	if matched := reUsername.MatchString(username); !matched {
		panic(ErrInvalidUsername)
	}

	registrant := runtime.PreviousRealm().Address()
	if err := susers.RegisterUser(cross, username, registrant); err != nil {
		panic(err)
	}

	latestUsers.Append(username)
	chain.Emit("Registration", "address", registrant.String(), "name", username)
}
