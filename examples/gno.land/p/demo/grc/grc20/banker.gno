package grc20

import (
	"std"
)

func PrevRealmBanker(b *Bank, slug string) *accountBanker {
	return &accountBanker{
		accountFn: func() account {
			account := newAccount(std.PrevRealm().Addr(), slug)
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
	account := newAccount(std.PrevRealm().Addr(), slug)
	return &accountBanker{
		accountFn: func() account { return account },
		bank:      b,
	}
}

func AdminBanker(b *Bank) *adminBanker {
	prev := std.PrevRealm().Addr()
	if b.owner != prev {
		panic("AdminBanker() can only be called by the Bank owner")
	}
	return &adminBanker{
		bank: b,
	}
}

func (b *Bank) MockBanker() { panic("not implemented") }

type accountBanker struct {
	accountFn func() account
	bank      *Bank
}

// var _ Token = (*token)(nil)
func (ab *accountBanker) GetName() string     { return ab.bank.name }
func (ab *accountBanker) GetSymbol() string   { return ab.bank.symbol }
func (ab *accountBanker) GetDecimals() uint   { return ab.bank.decimals }
func (ab *accountBanker) TotalSupply() uint64 { return ab.bank.totalSupply }

func (ab *accountBanker) BalanceOf(owner std.Address) uint64 {
	return ab.bank.BalanceOf(owner)
}

func (ab *accountBanker) Transfer(to std.Address, amount uint64) error {
	caller := std.PrevRealm().Addr()
	return ab.bank.Transfer(caller, to, amount)
}

func (ab *accountBanker) Allowance(owner, spender std.Address) uint64 {
	return ab.bank.Allowance(owner, spender)
}

func (ab *accountBanker) Approve(spender std.Address, amount uint64) error {
	caller := std.PrevRealm().Addr()
	return ab.bank.Approve(caller, spender, amount)
}

func (ab *accountBanker) TransferFrom(from, to std.Address, amount uint64) error {
	spender := std.PrevRealm().Addr()
	if err := ab.bank.SpendAllowance(from, spender, amount); err != nil {
		return err
	}
	return ab.bank.Transfer(from, to, amount)
}

type adminBanker struct {
	bank *Bank
}
