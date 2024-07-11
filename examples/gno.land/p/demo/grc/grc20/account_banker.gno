package grc20

import (
	"std"
)

func PrevRealmBanker(b *Bank) *fnBanker {
	return &fnBanker{
		accountFn: func() std.Address {
			caller := std.PrevRealm().Addr()
			return caller
		},
		bank: b,
	}
}

func ReadonlyBanker(b *Bank) *fnBanker {
	return &fnBanker{
		accountFn: nil,
		bank:      b,
	}
}

func AccountBanker(b *Bank, slug string) *fnBanker {
	caller := std.PrevRealm().Addr()
	account := AccountSlugAddr(caller, slug)
	return &fnBanker{
		accountFn: func() std.Address {
			return account
		},
		bank: b,
	}
}

// read-only methods

func (ab fnBanker) GetName() string                    { return ab.bank.name }
func (ab fnBanker) GetSymbol() string                  { return ab.bank.symbol }
func (ab fnBanker) GetDecimals() uint                  { return ab.bank.decimals }
func (ab fnBanker) TotalSupply() uint64                { return ab.bank.adm.totalSupply }
func (ab fnBanker) BalanceOf(owner std.Address) uint64 { return ab.bank.BalanceOf(owner) }
func (ab fnBanker) Allowance(owner, spender std.Address) uint64 {
	return ab.bank.Allowance(owner, spender)
}

// read-write methods

func (ab *fnBanker) Transfer(to std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	caller := ab.accountFn()
	return ab.bank.adm.Transfer(caller, to, amount)
}

func (ab *fnBanker) Approve(spender std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	caller := ab.accountFn()
	return ab.bank.adm.Approve(caller, spender, amount)
}

func (ab *fnBanker) TransferFrom(owner, to std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	spender := ab.accountFn()
	return ab.bank.adm.TransferFrom(owner, spender, to, amount)
}
