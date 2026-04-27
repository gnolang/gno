package namereg

import (
	"chain"
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
	registerPrice = int64(20_000_000) // 20 GNOT (anti-squatting deterrent)
	latestUsers   = fifo.New(100)     // Save the latest 100 users for rendering purposes
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
	//       this realm, leaving the envelope describing a phantom payment.
	//       IsUserCall is the only PreviousRealm shape where the tx-send
	//       envelope is guaranteed to have landed at this realm's address.
	//
	//   (b) cur.SentCoins() amount must exactly equal registerPrice.
	//       SentCoins() is the receiving-realm-scoped accessor: it returns
	//       the tx envelope ONLY when the calling realm is the first realm
	//       in the call stack. Structurally safer than banker.OriginSend(),
	//       which any realm in the chain can read.
	//
	// (a) is kept for explicit error messaging (ErrNonUserCall is clearer
	// than ErrInvalidPayment when the failure mode is "intermediate-realm
	// caller"). SentCoins() would also reject this case structurally.
	if !runtime.PreviousRealm().IsUserCall() {
		panic(ErrNonUserCall)
	}

	if paused {
		panic(ErrPaused)
	}

	// Walk the envelope to find ugnot. SentCoins() returns gnocoins
	// which doesn't expose an AmountOf helper.
	var ugnotAmount int64
	for _, c := range cur.SentCoins() {
		if c.Denom == "ugnot" {
			ugnotAmount = c.Amount
			break
		}
	}
	if ugnotAmount != registerPrice {
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
