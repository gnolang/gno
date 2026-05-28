// Package royalty is an EIP-2981 royalty extension for grc721.
//
// A PrivateLedger attaches per-token royalty info to a pre-constructed
// *grc721.Token. The ledger does not own the core token, so a single
// collection may carry multiple extensions (e.g. metadata and royalty) at
// the same time.
package royalty

import (
	"errors"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/bptree/v0"
)

// IGRC2981 follows EIP-2981.
type IGRC2981 interface {
	RoyaltyInfo(tid grc721.TokenID, salePrice int64) (address, int64, error)
}

// Teller is the royalty extension's read-only high-level API.
//
// Per EIP-2981 semantics, royalty configuration is a collection-level
// decision controlled by the issuing realm, not by individual token
// owners. Mutating writes therefore live on *PrivateLedger only; the
// host realm must guard SetTokenRoyalty behind its own admin policy.
// This Teller surfaces RoyaltyInfo as a safe-to-share read view.
type Teller interface {
	IGRC2981
}

// RoyaltyInfo records the per-token royalty configuration. Bps is the
// rate in basis points (1 bps = 0.01%, 10_000 bps = 100%) per EIP-2981.
type RoyaltyInfo struct {
	PaymentAddress address
	Bps            int64
}

// PrivateLedger holds per-token royalty state for the collection. It
// provides administrative functions for writing royalty entries.
//
// The PrivateLedger is not safe to expose publicly, as it allows direct,
// unrestricted access to all administrative functions.
type PrivateLedger struct {
	// Underlying core token; used to validate tid existence on writes.
	token     *grc721.Token
	royalties *bptree.BPTree
	maxBps    int64
}

var (
	ErrInvalidRoyaltyBps            = errors.New("invalid royalty bps")
	ErrInvalidRoyaltyPaymentAddress = errors.New("invalid royalty payment address")
	ErrInvalidMaxBps                = errors.New("invalid max bps")
)

const (
	// MaxBps is the maximum royalty representable in basis points (100%).
	MaxBps = 10_000

	// BpsDenominator is the denominator used to convert basis points to
	// a fraction. 10_000 bps = 100%.
	BpsDenominator = 10_000
)

// fnTeller is the canonical Teller implementation. accountFn supplies the
// acting principal for write methods; a nil accountFn marks the Teller as
// read-only.
type fnTeller struct {
	ledger    *PrivateLedger
	accountFn func(_ int, rlm realm) address
}

var (
	_ IGRC2981 = (*PrivateLedger)(nil)
	_ Teller   = (*fnTeller)(nil)
)

// IsCanonicalTeller reports whether t is the canonical *fnTeller produced by
// PrivateLedger.CallerTeller / RealmTeller / ReadonlyTeller /
// ImpersonateTeller. Use this at any public entry point that accepts a
// Teller from an external caller before invoking its methods.
//
// Foreign types — including embedding-based wrappers — are rejected because
// type assertions are nominal. Mirrors grc721.IsCanonicalTeller.
func IsCanonicalTeller(t Teller) bool {
	_, ok := t.(*fnTeller)
	return ok
}
