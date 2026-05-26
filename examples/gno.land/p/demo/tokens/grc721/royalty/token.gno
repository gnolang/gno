package royalty

import (
	"math/overflow"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/bptree/v0"
)

// NewPrivateLedger attaches a royalty ledger to an already-constructed core
// token. tok must be a non-nil *grc721.Token (typically returned from
// grc721.NewToken); the ledger holds a pointer to tok solely to validate tid
// existence on writes and does NOT own the core ledger.
//
// maxBps caps per-token royalties; passing 0 defaults to MaxBps.
//
// Typical call inside a realm's crossing function:
//
//	tok, coreLedger := grc721.NewToken(0, cur, "FOO", "FOO")
//	rLedger := royalty.NewPrivateLedger(tok, 1000)
//	mLedger := metadata.NewPrivateLedger(tok) // can coexist
//
// Do not export the returned *PrivateLedger.
func NewPrivateLedger(tok *grc721.Token, maxBps int64) *PrivateLedger {
	if tok == nil {
		panic("Token cannot be nil")
	}

	if maxBps == 0 {
		maxBps = MaxBps
	}

	if maxBps < 0 || maxBps > MaxBps {
		panic(ErrInvalidMaxBps)
	}

	return &PrivateLedger{
		token:     tok,
		royalties: bptree.NewBPTree32(),
		maxBps:    maxBps,
	}
}

// MaxBps returns the maximum royalty bps configured for this collection.
func (led *PrivateLedger) MaxBps() int64 { return led.maxBps }

// RoyaltyInfo returns the payment address and royalty amount due for a
// sale of tid at salePrice, satisfying IGRC2981.
func (led *PrivateLedger) RoyaltyInfo(tid grc721.TokenID, salePrice int64) (address, int64, error) {
	val, found := led.royalties.Get(tid.String())
	if !found {
		return "", 0, grc721.ErrInvalidTokenId
	}

	info := val.(RoyaltyInfo)
	amount := overflow.Mul64p(salePrice, info.Bps) / BpsDenominator
	return info.PaymentAddress, amount, nil
}

// Burn removes any royalty entry associated with tid from this extension
// ledger. It is the host realm's responsibility to call Burn on each
// attached extension ledger right after grc721.PrivateLedger.Burn so that
// stale per-tid state does not linger and resurface if the same tid is
// minted again.
//
// Burn is a no-op if tid has no royalty entry.
func (led *PrivateLedger) Burn(tid grc721.TokenID) {
	led.royalties.Remove(tid.String())
}

// SetTokenRoyalty stores royalty information for tid.
//
// Validity invariants:
//   - info.PaymentAddress is valid
//   - info.Bps is in [0, maxBps]
//   - tid exists on the underlying core token
func (led *PrivateLedger) SetTokenRoyalty(tid grc721.TokenID, info RoyaltyInfo) error {
	if !info.PaymentAddress.IsValid() {
		return ErrInvalidRoyaltyPaymentAddress
	}

	if info.Bps < 0 || info.Bps > led.maxBps {
		return ErrInvalidRoyaltyBps
	}

	if _, err := led.token.OwnerOf(tid); err != nil {
		return err
	}

	led.royalties.Set(tid.String(), info)
	return nil
}
