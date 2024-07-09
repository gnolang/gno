package grc20

import (
	"std"
)

func PrevRealmBanker(b *Bank, slug string) *accountBanker {
	return &accountBanker{
		accountFn: func() std.Address {
			prev := std.PrevRealm().Addr()
			account := AccountSlugAddr(prev, slug)
			return account
		},
		bank: b,
	}
}

func ReadonlyBanker(b *Bank, slug string) *accountBanker {
	return &accountBanker{
		accountFn: nil,
		bank:      b,
	}
}

func LocalBanker(b *Bank, slug string) *accountBanker {
	prev := std.PrevRealm().Addr()
	account := AccountSlugAddr(prev, slug)
	return &accountBanker{
		accountFn: func() std.Address {
			return account
		},
		bank: b,
	}
}

// var _ Token = (*token)(nil)
func (ab *accountBanker) GetName() string                    { return ab.bank.name }
func (ab *accountBanker) GetSymbol() string                  { return ab.bank.symbol }
func (ab *accountBanker) GetDecimals() uint                  { return ab.bank.decimals }
func (ab *accountBanker) TotalSupply() uint64                { return ab.bank.totalSupply_ }
func (ab *accountBanker) BalanceOf(owner std.Address) uint64 { return ab.bank.balanceOf(owner) }

func (ab *accountBanker) Transfer(to std.Address, amount uint64) error {
	caller := ab.accountFn()
	return ab.bank.transfer(caller, to, amount)
}

func (ab *accountBanker) Allowance(owner, spender std.Address) uint64 {
	return ab.bank.allowance(owner, spender)
}

func (ab *accountBanker) Approve(spender std.Address, amount uint64) error {
	caller := ab.accountFn()
	return ab.bank.approve(caller, spender, amount)
}

func (ab *accountBanker) TransferFrom(from, to std.Address, amount uint64) error {
	spender := ab.accountFn()
	if err := ab.bank.spendAllowance(from, spender, amount); err != nil {
		return err
	}
	return ab.bank.transfer(from, to, amount)
}
