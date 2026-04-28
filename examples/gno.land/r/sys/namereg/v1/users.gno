package namereg

import (
	"chain"
	"chain/banker"
	"chain/runtime"

	"gno.land/p/moul/fifo"
	susers "gno.land/r/sys/users"
)

// MinRegisterPrice is the lowest price (in ugnot) that
// ProposeNewRegisterPrice will accept. Set to 0 — registration is free
// by default; governance can raise the price via ProposeNewRegisterPrice
// without a floor.
const MinRegisterPrice = int64(0)

var (
	registerPrice = int64(0)      // free by default; governance can raise via ProposeNewRegisterPrice
	latestUsers   = fifo.New(100) // Save the latest 100 users for rendering purposes
)

// Register registers a new username for the caller.
//
// Valid usernames match `nym-[a-z]{5,13}\d{3}`:
//   - literal `nym-` prefix (4 chars)
//   - 5-13 lowercase letters (the alpha stem)
//   - exactly 3 trailing decimal digits
//
// Total length 12-20 chars. The alpha stem additionally must NOT start
// with `gno`/`gi`/`gl` and must not match a reserved role name (with
// implicit `s`-suffix expansion). See ValidateNymFormat for the
// format/blacklist check.
//
// Canonical-collision detection is enforced atomically by
// susers.RegisterUser via the unified canonical store in r/sys/users
// (decision: per Option B, every controller participates in the same
// canonical-form lookup keyed by full canonical name).
//
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

	// Format + prefix + reserved-name check. ValidateNymFormat returns
	// one of ErrInvalidFormat, ErrReservedPrefix, ErrBlacklisted.
	if err := ValidateNymFormat(username); err != nil {
		panic(err)
	}

	// Delegate the canonical-collision check + nameStore write atomically
	// to r/sys/users. Returns susers.ErrCanonicalCollision if the
	// canonical form clashes with an existing registration in any
	// controller.
	registrant := runtime.PreviousRealm().Address()
	if err := susers.RegisterUser(cross, username, registrant); err != nil {
		panic(err)
	}

	latestUsers.Append(username)
	chain.Emit("Registration", "address", registrant.String(), "name", username)
}
