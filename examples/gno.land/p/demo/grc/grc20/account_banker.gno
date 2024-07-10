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

// var _ Token = (*token)(nil)
func (ab *fnBanker) GetName() string                    { return ab.bank.name }
func (ab *fnBanker) GetSymbol() string                  { return ab.bank.symbol }
func (ab *fnBanker) GetDecimals() uint                  { return ab.bank.decimals }
func (ab *fnBanker) TotalSupply() uint64                { return ab.bank.totalSupply_ }
func (ab *fnBanker) BalanceOf(owner std.Address) uint64 { return ab.bank.BalanceOf(owner) }

func (ab *fnBanker) Transfer(to std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	caller := ab.accountFn()
	return ab.bank.Transfer(caller, to, amount)
}

func (ab *fnBanker) Allowance(owner, spender std.Address) uint64 {
	return ab.bank.Allowance(owner, spender)
}

func (ab *fnBanker) Approve(spender std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	caller := ab.accountFn()
	return ab.bank.Approve(caller, spender, amount)
}

func (ab *fnBanker) TransferFrom(from, to std.Address, amount uint64) error {
	if ab.accountFn == nil {
		return ErrReadonly
	}
	spender := ab.accountFn()
	if err := ab.bank.SpendAllowance(from, spender, amount); err != nil {
		return err
	}
	return ab.bank.Transfer(from, to, amount)
}
