package namereg

import (
	"chain"
	"chain/runtime/unsafe"

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
func Register(cur realm, username string) {
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
	if !cur.Previous().IsUserCall() {
		panic(ErrNonUserCall)
	}

	if paused {
		panic(ErrPaused)
	}

	sentCoins := unsafe.OriginSend()
	if sentCoins.AmountOf("ugnot") != registerPrice {
		panic(errInvalidPayment())
	}

	// Make sure no payment has been sent by mistake when registration is free
	if registerPrice == 0 && !sentCoins.IsZero() {
		panic(ErrNoPaymentExpected)
	}

	// Fee must only be payed in ugnot and no other coins must be sent by mistake
	if registerPrice > 0 && len(sentCoins) > 1 {
		panic(ErrInvalidPaymentCoin)
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
	registrant := cur.Previous().Address()
	if err := susers.RegisterUser(cross(cur), username, registrant); err != nil {
		panic(err)
	}

	latestUsers.Append(username)
	chain.Emit("Registration", "address", registrant.String(), "name", username)
}
