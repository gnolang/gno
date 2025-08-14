package mytoken

import (
	"std"

	"gno.land/p/demo/grc/grc20"
)

var (
	token *grc20.Token
	ledger *grc20.PrivateLedger
)

func init() {
	token, ledger = grc20.NewToken("MyToken", "MTK", 6)

	admin := std.CurrentRealm().Address()
	_ = ledger.Mint(admin, 1_000_000*1_000_000)
}

func Name() string {
    return token.GetName()
}

func Symbol() string {
    return token.GetSymbol()
}

func Decimals() int {
    return token.GetDecimals()
}

func TotalSupply() int64 {
    return token.TotalSupply()
}

func BalanceOf(addr std.Address) int64 {
    return token.BalanceOf(addr)
}

func Transfer(to std.Address, amount int64) error {
    teller := token.CallerTeller()
    return teller.Transfer(to, amount)
}

func Approve(spender std.Address, amount int64) error {
    teller := token.CallerTeller()
    return teller.Approve(spender, amount)
}

func TransferFrom(owner, to std.Address, amount int64) error {
    teller := token.CallerTeller()
    return teller.TransferFrom(owner, to, amount)
}

func RenderDemo() string {
	out := "## Live Demo\n"
    out += token.RenderHome()
    return out
}
