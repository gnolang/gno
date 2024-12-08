package mytoken

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
)

// TotalSupply returns the total supply of mytoken
func TotalSupply() uint64 {
	return mytoken.TotalSupply()
}

// Decimals returns the number of decimals of mytoken
func Decimals() uint {
	return mytoken.GetDecimals()
}

// BalanceOf returns the balance mytoken for `account`
func BalanceOf(account std.Address) uint64 {
	return mytoken.BalanceOf(account)
}

// Allowance returns the allowance of spender on owner's balance
func Allowance(owner, spender std.Address) uint64 {
	return mytoken.Allowance(owner, spender)
}

// Transfer transfers amount from caller to recipient
func Transfer(recipient std.Address, amount uint64) {
	checkErr(mytoken.Transfer(recipient, amount))
}

// Approve approves amount of caller's tokens to be spent by spender
func Approve(spender std.Address, amount uint64) {
	checkErr(mytoken.Approve(spender, amount))
}

// TransferFrom transfers `amount` of tokens from `from` to `to`
func TransferFrom(from, to std.Address, amount uint64) {
	checkErr(mytoken.TransferFrom(from, to, amount))
}

// Mint mints amount of tokens to address. Callable only by admin of token
func Mint(address std.Address, amount uint64) {
	assertIsAdmin(std.PrevRealm().Addr())
	checkErr(banker.Mint(address, amount))
}

// Burn burns amount of tokens from address. Callable only by admin of token
func Burn(address std.Address, amount uint64) {
	assertIsAdmin(std.PrevRealm().Addr())
	checkErr(banker.Burn(address, amount))
}

// Render renders the state of the realm
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		// Default GRC20 render
		return mytoken.RenderHome()
	case c == 2 && parts[0] == "balance":
		// Render balance of specific address
		owner := std.Address(parts[1])
		balance, _ := mytoken.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

// assertIsAdmin asserts the address is the admin of token
func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}

// checkErr asserts the function didn't returned an error
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
