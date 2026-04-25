package namereg

import (
	"chain"
	"chain/banker"
	"chain/runtime"

	"gno.land/p/moul/fifo"
	susers "gno.land/r/sys/users"
)

// MinRegisterPrice is the lowest price (in ugnot) that
// ProposeNewRegisterPrice will accept. Set to 0.1 GNOT — chosen so the
// default has 3 orders of magnitude of governance headroom for lowering
// the price without dropping all the way to zero. (audit finding #14)
const MinRegisterPrice = int64(100_000) // 0.1 GNOT

var (
	registerPrice = int64(200_000_000) // 200 GNOT (anti-squatting deterrent)
	latestUsers   = fifo.New(100)      // Save the latest 100 users for rendering purposes
)

// Register registers a new username for the caller.
//
// Valid usernames match `nym-[a-z]{5,13}\d{3}`:
//   - literal `nym-` prefix (4 chars)
//   - 5-13 lowercase letters (the alpha stem)
//   - exactly 3 trailing decimal digits
//
// Total length 12-20 chars. The alpha stem additionally must NOT start
// with `gno`/`gi`/`gl`, must not match a reserved role name (with
// implicit `s`-suffix expansion), and must not canonicalize to an
// alpha stem already registered through this controller (`l→i`
// substitution). See ValidateNymFormat and IsStemTaken for the
// individual checks; both are exported for use by future controllers.
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

	// Canonical-collision check. Stem is everything between the `nym-`
	// prefix and the 3 trailing digits.
	stem := username[4 : len(username)-3]
	if _, taken := IsStemTaken(stem); taken {
		panic(ErrCanonicalCollision)
	}
	canonical := Canonicalize(stem)

	registrant := runtime.PreviousRealm().Address()
	if err := susers.RegisterUser(cross, username, registrant); err != nil {
		panic(err)
	}

	// Persist the canonical→original mapping so future registrations
	// can detect collisions. Stored as the bare alpha stem (not the
	// full nym-XXX999) because the prefix and digits are constant
	// across all Open Nym Tier names and contribute nothing to distinction.
	canonicalStore.Set(canonical, stem)

	latestUsers.Append(username)
	chain.Emit("Registration", "address", registrant.String(), "name", username)
}
