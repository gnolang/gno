package grc20

import (
	"std"
)

func PrevRealmTeller(b *Token) GRC20 {
	if b == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: func() std.Address {
			caller := std.PrevRealm().Addr()
			return caller
		},
		Token: b,
	}
}

func RealmTeller(b *Token) GRC20 {
	if b == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: func() std.Address {
			caller := std.CurrentRealm().Addr()
			return caller
		},
		Token: b,
	}
}

func ReadonlyTeller(b *Token) GRC20 {
	if b == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: nil,
		Token:     b,
	}
}

func SubAccountTeller(b *Token, slug string) GRC20 {
	if b == nil {
		panic("Token cannot be nil")
	}

	caller := std.CurrentRealm().Addr()
	account := accountSlugAddr(caller, slug)

	return &fnTeller{
		accountFn: func() std.Address {
			return account
		},
		Token: b,
	}
}

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

// accountSlugAddr returns the address derived from the specified address and slug.
func accountSlugAddr(addr std.Address, slug string) std.Address {
	// XXX: use a new `std.XXX` call for this.
	if slug == "" {
		return addr
	}
	key := addr.String() + "/" + slug
	return std.DerivePkgAddr(key) // temporarily using this helper
}
