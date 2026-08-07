package grc721

import (
	"chain"
)

// Each write resolves the acting account as rlm.Previous() — the realm that crossed in.
func (tok *Token) CallerTeller() Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: func(_ int, rlm realm) address {
			return rlm.Previous().Address()
		},
		Token: tok,
	}
}

// Write methods return ErrReadonly.
func (tok *Token) ReadonlyTeller() Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: nil,
		Token:     tok,
	}
}

// Permanently acts as the calling realm (verified via IsCurrent).
func (tok *Token) RealmTeller(_ int, rlm realm) Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	if !rlm.IsCurrent() {
		panic(ErrSpoofedRealm)
	}

	caller := rlm.Address()

	return &fnTeller{
		accountFn: func(_ int, _ realm) address {
			return caller
		},
		Token: tok,
	}
}

// Like RealmTeller but acts as a sub-account derived from slug.
func (tok *Token) RealmSubTeller(_ int, rlm realm, slug string) Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	if !rlm.IsCurrent() {
		panic(ErrSpoofedRealm)
	}

	account := accountSlugAddr(rlm.Address(), slug)

	return &fnTeller{
		accountFn: func(_ int, _ realm) address {
			return account
		},
		Token: tok,
	}
}

// Admin-grade: issuer-only (holds the PrivateLedger). Panics on invalid addr to close the empty-sentinel hole.
func (led *PrivateLedger) ImpersonateTeller(addr address) Teller {
	if led == nil {
		panic("Ledger cannot be nil")
	}

	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	return &fnTeller{
		accountFn: func(_ int, _ realm) address {
			return addr
		},
		Token: led.token,
	}
}

func (ft *fnTeller) Approve(_ int, rlm realm, to address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	caller := ft.accountFn(0, rlm)

	return ft.Token.ledger.Approve(caller, to, tid)
}

func (ft *fnTeller) SetApprovalForAll(_ int, rlm realm, operator address, approved bool) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	caller := ft.accountFn(0, rlm)

	return ft.Token.ledger.SetApprovalForAll(caller, operator, approved)
}

func (ft *fnTeller) TransferFrom(_ int, rlm realm, from, to address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	spender := ft.accountFn(0, rlm)

	return ft.Token.ledger.TransferFrom(spender, from, to, tid)
}

func accountSlugAddr(addr address, slug string) address {
	if slug == "" {
		return addr
	}

	key := addr.String() + "/" + slug

	return chain.PackageAddress(key)
}
