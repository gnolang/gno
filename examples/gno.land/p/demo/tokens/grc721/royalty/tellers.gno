package royalty

import (
	"gno.land/p/demo/tokens/grc721"
)

// CallerTeller returns a royalty Teller that resolves the acting principal
// from rlm.Previous() per write call. Safe to expose publicly.
func (led *PrivateLedger) CallerTeller() Teller {
	if led == nil {
		panic("Ledger cannot be nil")
	}

	return &fnTeller{
		ledger: led,
		accountFn: func(_ int, rlm realm) address {
			return rlm.Previous().Address()
		},
	}
}

// ReadonlyTeller returns a royalty Teller whose write methods all return
// grc721.ErrReadonly. Read methods pass through unchanged.
func (led *PrivateLedger) ReadonlyTeller() Teller {
	if led == nil {
		panic("Ledger cannot be nil")
	}

	return &fnTeller{
		ledger:    led,
		accountFn: nil,
	}
}

// RealmTeller captures rlm.Address() at construction.
//
// rlm must be the caller's own captured cur (asserted via rlm.IsCurrent()).
// The address is frozen eagerly at construction.
func (led *PrivateLedger) RealmTeller(_ int, rlm realm) Teller {
	if led == nil {
		panic("Ledger cannot be nil")
	}
	if !rlm.IsCurrent() {
		panic(grc721.ErrSpoofedRealm)
	}

	caller := rlm.Address()

	return &fnTeller{
		ledger:    led,
		accountFn: func(_ int, _ realm) address { return caller },
	}
}

// ImpersonateTeller returns a royalty Teller acting as addr. Reserved for
// admin paths; possessing *PrivateLedger is the authorization gate.
func (led *PrivateLedger) ImpersonateTeller(addr address) Teller {
	if led == nil {
		panic("Ledger cannot be nil")
	}

	if !addr.IsValid() {
		panic(grc721.ErrInvalidAddress)
	}

	return &fnTeller{
		ledger:    led,
		accountFn: func(_ int, _ realm) address { return addr },
	}
}

// generic tellers methods.
//

// RoyaltyInfo is a read-only delegation; no caller check needed.
func (ft *fnTeller) RoyaltyInfo(tid grc721.TokenID, salePrice int64) (address, int64, error) {
	return ft.ledger.RoyaltyInfo(tid, salePrice)
}
