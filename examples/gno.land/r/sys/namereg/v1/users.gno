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

// MinRegisterPrice is the lowest price (in ugnot) that
// ProposeNewRegisterPrice will accept. Set to 0.1 GNOT — chosen so the
// default has 3 orders of magnitude of governance headroom for lowering
// the price without dropping all the way to zero. (audit finding #14)
const MinRegisterPrice = int64(100_000) // 0.1 GNOT

var (
	registerPrice = int64(200_000_000) // 200 GNOT (anti-squatting deterrent)
	latestUsers   = fifo.New(10)       // Save the latest 10 users for rendering purposes
	reUsername    = regexp.MustCompile(reValidUsername)
)

// Register registers a new username for the caller.
// A valid username must start with a minimum of 3 letters,
// end with a minimum of 3 numbers, and be less than 20 chars long.
// All letters must be lowercase, and the only valid special char is `_`.
// Only direct EOA (maketx call) invocations are supported.
func Register(_ realm, username string) {
	// Anti-squatting payment check, two paired guards:
	//
	//   (a) PreviousRealm must be a pure EOA (IsUserCall: pkgPath == "").
	//       This excludes intermediate code realms AND user-run ephemeral
	//       realms ("maketx run" scripts). Both can attach -send to the
	//       tx but spend the coins on something other than forwarding to
	//       this realm, leaving OriginSend() describing a phantom payment.
	//       IsUserCall is the only PreviousRealm shape where the tx-send
	//       envelope is guaranteed to have landed at this realm's address.
	//
	//   (b) OriginSend amount must exactly equal registerPrice. Verifies
	//       the tx actually attached the expected amount.
	//
	// Both checks MUST run together. Removing (a) alone makes (b) meaningless
	// because OriginSend() describes tx intent, not realm receipt.
	if !runtime.PreviousRealm().IsUserCall() {
		panic(ErrNonUserCall)
	}

	if paused {
		panic(ErrPaused)
	}

	if banker.OriginSend().AmountOf("ugnot") != registerPrice {
		panic(errInvalidPayment())
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
