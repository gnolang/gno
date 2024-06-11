// Package grc20factory defines a factory realm for grc20 tokens.
// It serves as a registry/proxy for accessing all tokens registered in it.

package grc20factory

import (
	"bytes"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

type tokenWithAdmin struct {
	token    *grc20.Token
	admin    std.Address
	fauceted bool
}

var tokens avl.Tree // symbol -> tokenWithAdmin

func NewToken(name, symbol string, decimals uint8, initialMint uint64) {
	admin := std.PrevRealm().Addr()
	NewTokenWithAdmin(name, symbol, decimals, initialMint, admin)
}

func NewTokenWithAdmin(name, symbol string, decimals uint8, initialMint uint64, admin std.Address) {
	exists := tokens.Has(symbol)
	if exists {
		err := ufmt.Sprintf("token with symbol $%s already exists", symbol)
		panic(err)
	}

	newToken := grc20.NewGRC20Token(name, symbol, decimals)
	newToken.Mint(admin, initialMint)

	t := tokenWithAdmin{
		token:    newToken,
		admin:    admin,
		fauceted: false, // faucet disabled by default
	}

	tokens.Set(symbol, &t)
}

// Getters

// TotalSupply gets the total supply of a token specified by symbol
func TotalSupply(symbol string) uint64 {
	t := mustTokenBySymbol(symbol)
	return t.token.TotalSupply()
}

func BalanceOf(symbol string, owner std.Address) uint64 {
	t := mustTokenBySymbol(symbol)
	return t.token.BalanceOf(owner)
}

func Allowance(symbol string, owner, spender std.Address) uint64 {
	t := mustTokenBySymbol(symbol)
	return t.token.Allowance(owner, spender)
}

// Setters

func Transfer(symbol string, to std.Address, amount uint64) {
	t := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()

	t.token.Transfer(to, amount)
}

func Approve(symbol string, spender std.Address, amount uint64) {
	t := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()

	t.token.Approve(spender, amount)
}

func TransferFrom(symbol string, from, to std.Address, amount uint64) {
	t := mustTokenBySymbol(symbol)
	t.token.TransferFrom(from, to, amount)
}

// Faucet

func Faucet(symbol string) {
	t := mustTokenBySymbol(symbol)
	if !t.fauceted {
		err := ufmt.Sprintf("faucet is paused for token %s ($%s)", t.token.Name(), t.token.Symbol())
		panic(err)
	}

	// Mint 1000 tokens to caller
	caller := std.PrevRealm().Addr()
	t.token.Mint(caller, 1000*10*uint64(t.token.Decimals()))
}

// Administration

func Mint(symbol string, address std.Address, amount uint64) {
	t := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller, t.admin)

	t.token.Mint(address, amount)
}

func Burn(symbol string, address std.Address, amount uint64) {
	t := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller, t.admin)

	t.token.Burn(address, amount)
}

func EnableFaucet(symbol string) {
	t := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller, t.admin)

	t.fauceted = true
}

func DisableFaucet(symbol string) {
	t := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller, t.admin)

	t.fauceted = false
}

// Rendering
//

func renderList() string {
	var buf bytes.Buffer

	buf.WriteString(ufmt.Sprintf("# GRC20 Token Factory\n\n"))

	tokens.Iterate("", "", func(key string, value interface{}) bool {
		t := value.(grc20.Token)
		buf.WriteString(ufmt.Sprintf("[Token $%s](r/demo/grc20factory:%s)\n\n", t.Symbol()))

		return false
	})

	return buf.String()
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "": // pkgpath
		return renderList()
	case c == 1: // pkgpath:symbol
		symbol := parts[0]
		t := mustTokenBySymbol(symbol)
		return t.token.RenderHome()
	case c == 3 && parts[1] == "balance": // pkgpath:symbol/balance/address_or_username
		symbol := parts[0]

		t := mustTokenBySymbol(symbol)
		owner := std.Address(parts[2])

		return strconv.FormatUint(t.token.BalanceOf(owner), 10)
	default:
		return "404\n"
	}
}

// helpers.
//

func assertIsAdmin(caller, admin std.Address) {
	if caller != admin {
		panic("restricted access")
	}
}

func mustTokenBySymbol(symbol string) *tokenWithAdmin {
	t, exists := tokens.Get(symbol)
	if !exists {
		panic("token does not exist")
	}

	return t.(*tokenWithAdmin)
}
