package oo

import (
	"std"

	"gno.land/p/demo/grc/grc20"
)

type OOT struct {
	*grc20.Token
	*grc20.PrivateLedger
}

func newOOToken(name, symbol string, decimals int) *OOT {
	token := &OOT{}
	t, adm := grc20.NewToken(name, symbol, decimals)
	token.Token = t
	token.PrivateLedger = adm
	return token
}

// Mint create new tokens to an address.
func (t *OOT) mint(to std.Address, amount int64) {
	t.PrivateLedger.Mint(to, amount)
}

// Burn destroy tokens from an address.
func (t *OOT) burn(from std.Address, amount int64) {
	t.PrivateLedger.Burn(from, amount)
}

func (t *OOT) balanceOf(addr std.Address) int64 {
	return t.Token.BalanceOf(addr)
}
