package grc20

import (
	"std"
)

// CallerTeller returns a GRC20 compatible teller that checks the PrevRealm caller for each call.
// It's usually safe to expose it publicly to let users manipulate their tokens directly, or for reams to use their allowance.
func (tok *Token) CallerTeller() GRC20 {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: func() std.Address {
			caller := std.PrevRealm().Addr()
			return caller
		},
		Token: tok,
	}
}

// ReadonlyTeller is a GRC20 compatible teller that panics for any write operation.
func (tok *Token) ReadonlyTeller() GRC20 {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: nil,
		Token:     tok,
	}
}

// RealmTeller returns a GRC20 compatible teller that will store the CurrentRealm caller and then reuse the stored caller for each call.
// It allows a realm contract to interact with a GRC20 with its own account.
// The initializer of this teller should usually never share a pointer to its RealmTeller except maybe for advanced delegation flows such as a DAO treasury management.
func (tok *Token) RealmTeller() GRC20 {
	if tok == nil {
		panic("Token cannot be nil")
	}

	caller := std.PrevRealm().Addr()

	return &fnTeller{
		accountFn: func() std.Address {
			return caller
		},
		Token: tok,
	}
}

// RealmSubTeller is like RealmTeller but uses the provided slug to derivate a subaccount.
func (tok *Token) RealmSubTeller(slug string) GRC20 {
	if tok == nil {
		panic("Token cannot be nil")
	}

	caller := std.PrevRealm().Addr()
	account := accountSlugAddr(caller, slug)

	return &fnTeller{
		accountFn: func() std.Address {
			return account
		},
		Token: tok,
	}
}

func (ledger *PrivateLedger) MasqueradingTeller(addr std.Address) GRC20 {
	if ledger == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: func() std.Address {
			return addr
		},
		Token: ledger.token,
	}
}

// generic tellers methods.
//

func (ft *fnTeller) Transfer(to std.Address, amount uint64) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}
	caller := ft.accountFn()
	return ft.Token.ledger.Transfer(caller, to, amount)
}

func (ft *fnTeller) Approve(spender std.Address, amount uint64) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}
	caller := ft.accountFn()
	return ft.Token.ledger.Approve(caller, spender, amount)
}

func (ft *fnTeller) TransferFrom(owner, to std.Address, amount uint64) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}
	spender := ft.accountFn()
	return ft.Token.ledger.TransferFrom(owner, spender, to, amount)
}

// helpers
//

// accountSlugAddr returns the address derived from the specified address and slug.
func accountSlugAddr(addr std.Address, slug string) std.Address {
	// XXX: use a new `std.XXX` call for this.
	if slug == "" {
		return addr
	}
	key := addr.String() + "/" + slug
	return std.DerivePkgAddr(key) // temporarily using this helper
}
