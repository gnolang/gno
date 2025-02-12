package qux

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"

	"gno.land/r/demo/grc20reg"
)

var (
	Token, privateLedger = grc20.NewToken("Qux", "QUX", 6)
	UserTeller           = Token.CallerTeller()
	owner                = ownable.NewWithAddress("g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d") // ADMIN
)

func init() {
	privateLedger.Mint(owner.Owner(), 100_000_000_000_000)
	getter := func() *grc20.Token { return Token }
	grc20reg.Register(getter, "")
}

func TotalSupply() uint64 {
	return UserTeller.TotalSupply()
}

func BalanceOf(owner std.Address) uint64 {
	return UserTeller.BalanceOf(owner)
}

func Allowance(owner, spender std.Address) uint64 {
	return UserTeller.Allowance(owner, spender)
}

func Transfer(to std.Address, amount uint64) {
	checkErr(UserTeller.Transfer(to, amount))
}

func Approve(spender std.Address, amount uint64) {
	checkErr(UserTeller.Approve(spender, amount))
}

func TransferFrom(from, to std.Address, amount uint64) {
	checkErr(UserTeller.TransferFrom(from, to, amount))
}

func Burn(from std.Address, amount uint64) {
	owner.AssertCallerIsOwner()
	checkErr(privateLedger.Burn(from, amount))
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return Token.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := std.Address(parts[1])
		balance := UserTeller.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
