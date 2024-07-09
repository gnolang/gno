package grc20

import (
	"std"
)

func AdminBanker(b *Bank) *adminBanker {
	prev := std.PrevRealm().Addr()
	if b.owner != prev {
		panic("AdminBanker() can only be called by the Bank owner")
	}
	return &adminBanker{
		bank: b,
	}
}

func (ab adminBanker) GetName() string     { return ab.bank.name }
func (ab adminBanker) GetSymbol() string   { return ab.bank.symbol }
func (ab adminBanker) GetDecimals() uint   { return ab.bank.decimals }
func (ab adminBanker) TotalSupply() uint64 { return ab.bank.totalSupply_ }
func (ab adminBanker) KnownAccounts() int  { return ab.bank.balances.Size() }
func (ab *adminBanker) Mint(address std.Address, amount uint64) error {
	return ab.bank.mint(address, amount)
}

func (ab *adminBanker) Burn(address std.Address, amount uint64) error {
	return ab.bank.burn(address, amount)
}

func (ab adminBanker) BalanceOf(address std.Address) uint64 {
	return ab.bank.balanceOf(address)
}

func (ab *adminBanker) SpendAllowance(owner, spender std.Address, amount uint64) error {
	return ab.bank.spendAllowance(owner, spender, amount)
}

func (ab *adminBanker) Transfer(from, to std.Address, amount uint64) error {
	return ab.bank.transfer(from, to, amount)
}

func (ab *adminBanker) TransferFrom(spender, from, to std.Address, amount uint64) error {
	return ab.bank.transferFrom(spender, from, to, amount)
}

func (ab *adminBanker) Allowance(owner, spender std.Address) uint64 {
	return ab.bank.allowance(owner, spender)
}

func (ab *adminBanker) Approve(owner, spender std.Address, amount uint64) error {
	return ab.bank.approve(owner, spender, amount)
}
