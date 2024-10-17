package grc20

import (
	"std"
)

func PrevRealmBanker(b *Bank) GRC20 {
	return &fnBanker{
		accountFn: func() std.Address {
			caller := std.PrevRealm().Addr()
			return caller
		},
		Bank: b,
	}
}

func ReadonlyBanker(b *Bank) GRC20 {
	return &fnBanker{
		accountFn: nil,
		Bank:      b,
	}
}

func AccountBanker(b *Bank, slug string) GRC20 {
	caller := std.CurrentRealm().Addr()
	account := AccountSlugAddr(caller, slug)
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
