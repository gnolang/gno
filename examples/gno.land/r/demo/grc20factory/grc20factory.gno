package foo20

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/r/demo/grc20reg"
)

// XXX: the p/grc20 library needs a refactor to change names (i.e., adminToken)

func init() {
	NewToken("lorem ipsum", "LOREM", 4, 0xcafebabe)
}

type token struct {
	adminToken *grc20.AdminToken
	admin      std.Address
}

var tokens avl.Tree // symbol -> token

func NewToken(name, symbol string, decimals uint, initialMint uint64) {
	admin := std.PrevRealm().Addr()
	NewTokenWithAdmin(name, symbol, decimals, initialMint, admin)
}

func NewTokenWithAdmin(name, symbol string, decimals uint, initialMint uint64, admin std.Address) {
	exists := tokens.Has(symbol)
	if exists {
		panic("token already exists")
	}

	newToken := grc20.NewAdminToken(name, symbol, decimals)
	newToken.Mint(admin, initialMint)
	t := token{
		adminToken: newToken,
		admin:      admin,
	}
	tokens.Set(symbol, &t)
	grc20reg.Register(newToken.GRC20(), symbol)
}

// method proxies as public functions.
//

// getters.

func TotalSupply(symbol string) uint64 {
	token := mustTokenBySymbol(symbol)
	return token.adminToken.TotalSupply()
}

func BalanceOf(symbol string, owner std.Address) uint64 {
	token := mustTokenBySymbol(symbol)
	balance, err := token.adminToken.BalanceOf(owner)
	checkErr(err)
	return balance
}

func Allowance(symbol string, owner, spender std.Address) uint64 {
	token := mustTokenBySymbol(symbol)
	allowance, err := token.adminToken.Allowance(owner, spender)
	checkErr(err)
	return allowance
}

// setters.

func Transfer(symbol string, to std.Address, amount uint64) {
	token := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	err := token.adminToken.Transfer(caller, to, amount)
	checkErr(err)
}

func Approve(symbol string, spender std.Address, amount uint64) {
	token := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	err := token.adminToken.Approve(caller, spender, amount)
	checkErr(err)
}

func TransferFrom(symbol string, from, to std.Address, amount uint64) {
	token := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	err := token.adminToken.TransferFrom(caller, from, to, amount)
	if err != nil {
		panic(err)
	}
}

// faucet.
func Faucet(symbol string) {
	token := mustTokenBySymbol(symbol)
	// FIXME: add limits?
	// FIXME: add payment in gnot?
	caller := std.PrevRealm().Addr()
	err := token.adminToken.Mint(caller, 1000*10000) // 1k
	checkErr(err)
}

// administration.

func Mint(symbol string, address std.Address, amount uint64) {
	token := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller, token.admin)
	err := token.adminToken.Mint(address, amount)
	checkErr(err)
}

func Burn(symbol string, address std.Address, amount uint64) {
	token := mustTokenBySymbol(symbol)
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller, token.admin)
	err := token.adminToken.Burn(address, amount)
	checkErr(err)
}

// render.
//

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return "TODO: list existing tokens and admins"
	case c == 1:
		symbol := parts[0]
		token := mustTokenBySymbol(symbol)
		return token.adminToken.RenderHome()
	case c == 3 && parts[1] == "balance":
		symbol := parts[0]
		token := mustTokenBySymbol(symbol)
		owner := std.Address(parts[2])
		balance, _ := token.adminToken.BalanceOf(owner)
		return strconv.FormatUint(balance, 10)
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

func mustTokenBySymbol(symbol string) *token {
	t, exists := tokens.Get(symbol)
	if !exists {
		panic("token does not exist")
	}
	return t.(*token)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
