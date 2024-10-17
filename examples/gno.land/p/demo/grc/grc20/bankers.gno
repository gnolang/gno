package grc20

import (
	"std"
)

func PrevRealmBanker(b *Bank) GRC20 {
	if b == nil {
		panic("Bank cannot be nil")
	}

	return &fnBanker{
		accountFn: func() std.Address {
			caller := std.PrevRealm().Addr()
			return caller
		},
		Bank: b,
	}
}

func RealmBanker(b *Bank) GRC20 {
	if b == nil {
		panic("Bank cannot be nil")
	}

	return &fnBanker{
		accountFn: func() std.Address {
			caller := std.CurrentRealm().Addr()
			return caller
		},
		Bank: b,
	}
}

func ReadonlyBanker(b *Bank) GRC20 {
	if b == nil {
		panic("Bank cannot be nil")
	}

	return &fnBanker{
		accountFn: nil,
		Bank:      b,
	}
}

func SubAccountBanker(b *Bank, slug string) GRC20 {
	if b == nil {
		panic("Bank cannot be nil")
	}

	caller := std.CurrentRealm().Addr()
	account := accountSlugAddr(caller, slug)

	return &fnBanker{
		accountFn: func() std.Address {
			return account
		},
		Bank: b,
	}
}

func (ab *fnBanker) Transfer(to std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	caller := ab.accountFn()
	return ab.Bank.adm.Transfer(caller, to, amount)
}

func (ab *fnBanker) Approve(spender std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	caller := ab.accountFn()
	return ab.Bank.adm.Approve(caller, spender, amount)
}

func (ab *fnBanker) TransferFrom(owner, to std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	spender := ab.accountFn()
	return ab.Bank.adm.TransferFrom(owner, spender, to, amount)
}
