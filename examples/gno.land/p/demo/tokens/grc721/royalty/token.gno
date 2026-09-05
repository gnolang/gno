package royalty

import (
	"chain"
	"math"
	"strconv"

	"gno.land/p/demo/tokens/grc721"
)

const Kind = "royalty"

// NewRoyalty attaches a royalty extension; maxBps caps every rate and must be 0..10000.
func NewRoyalty(coreLedger *grc721.PrivateLedger, maxBps int64) (*Royalty, *Ledger) {
	if coreLedger == nil {
		panic("royalty: nil core ledger")
	}

	if maxBps < 0 || maxBps > FeeDenominator {
		panic(ErrMaxBpsRange)
	}

	st := &storage{maxBps: maxBps}
	core := coreLedger.ReadToken()
	led := &Ledger{core: core, st: st}

	coreLedger.RegisterExtension(led)

	return &Royalty{core: core, st: st}, led
}

func (r *Royalty) ExtensionKind() string { return Kind }

func (r *Royalty) TokenID() string { return r.core.ID() }

func (r *Royalty) MaxBps() int64 { return r.st.maxBps }

// RoyaltyInfo implements EIP-2981; per-token override takes precedence over the
// default, and no royalty returns (zeroAddress, 0) — the "no royalty" signal.
func (r *Royalty) RoyaltyInfo(tid grc721.TokenID, salePrice int64) (address, int64, error) {
	if salePrice < 0 {
		return zeroAddress, 0, ErrInvalidSalePrice
	}

	info, ok := r.resolve(tid)
	if !ok {
		return zeroAddress, 0, nil
	}

	if info.Bps != 0 && salePrice > math.MaxInt64/info.Bps {
		return zeroAddress, 0, ErrInvalidSalePrice
	}

	amount := salePrice * info.Bps / FeeDenominator

	return info.Receiver, amount, nil
}

func (r *Royalty) DefaultRoyalty() (RoyaltyInfo, bool) {
	if !r.st.hasDefault {
		return RoyaltyInfo{}, false
	}

	return r.st.def, true
}

func (r *Royalty) TokenRoyalty(tid grc721.TokenID) (RoyaltyInfo, bool) {
	v := r.st.perToken.Get(tid.String())
	if v == nil {
		return RoyaltyInfo{}, false
	}

	return v.(RoyaltyInfo), true
}

// resolve returns the effective royalty: per-token override first, then default.
func (r *Royalty) resolve(tid grc721.TokenID) (RoyaltyInfo, bool) {
	if v := r.st.perToken.Get(tid.String()); v != nil {
		return v.(RoyaltyInfo), true
	}

	if r.st.hasDefault {
		return r.st.def, true
	}

	return RoyaltyInfo{}, false
}

// SetDefaultRoyalty sets the collection default for tokens without an override; issuer-only.
func (led *Ledger) SetDefaultRoyalty(receiver address, bps int64) error {
	if err := led.validate(receiver, bps); err != nil {
		return err
	}

	led.st.def = RoyaltyInfo{Receiver: receiver, Bps: bps}
	led.st.hasDefault = true

	led.emitRoyaltyUpdate("default", "", receiver, bps)
	return nil
}

func (led *Ledger) DeleteDefaultRoyalty() {
	led.st.hasDefault = false
	led.st.def = RoyaltyInfo{}

	led.emitRoyaltyUpdate("default", "", zeroAddress, 0)
}

// SetTokenRoyalty sets a per-token override for tid; issuer-only.
func (led *Ledger) SetTokenRoyalty(tid grc721.TokenID, receiver address, bps int64) error {
	if _, err := led.core.OwnerOf(tid); err != nil {
		return ErrTokenNotMinted
	}

	if err := led.validate(receiver, bps); err != nil {
		return err
	}

	led.st.perToken.Set(tid.String(), RoyaltyInfo{Receiver: receiver, Bps: bps})

	led.emitRoyaltyUpdate("token", tid.String(), receiver, bps)
	return nil
}

// DeleteTokenRoyalty clears the per-token override, falling back to the default.
func (led *Ledger) DeleteTokenRoyalty(tid grc721.TokenID) {
	led.st.perToken.Remove(tid.String())

	led.emitRoyaltyUpdate("token", tid.String(), zeroAddress, 0)
}

// emitRoyaltyUpdate signals a royalty change (EIP-2981 defines no events); a
// cleared royalty is reported as an empty receiver and bps 0.
func (led *Ledger) emitRoyaltyUpdate(scope, tokenID string, receiver address, bps int64) {
	chain.Emit(
		RoyaltyUpdateEvent,
		"token", led.core.ID(),
		"scope", scope,
		"tokenId", tokenID,
		"receiver", receiver.String(),
		"bps", strconv.FormatInt(bps, 10),
	)
}

func (led *Ledger) validate(receiver address, bps int64) error {
	if !receiver.IsValid() {
		return ErrInvalidReceiver
	}

	if bps < 0 || bps > led.st.maxBps {
		return ErrInvalidBps
	}

	return nil
}

func (led *Ledger) ExtensionKind() string { return Kind }

func (led *Ledger) OnMint(to address, tid grc721.TokenID)           {}
func (led *Ledger) OnTransfer(from, to address, tid grc721.TokenID) {}

// OnBurn clears the per-token override (cannot resurface on re-mint) but keeps
// the default, which is collection-level policy, not per-token state.
func (led *Ledger) OnBurn(tid grc721.TokenID) {
	led.st.perToken.Remove(tid.String())
}
