package wugnot

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/grc20reg"
)

var Token, adm = grc20.NewToken("wrapped GNOT", "wugnot", 0)

const (
	ugnotMinDeposit  uint64 = 1000
	wugnotMinDeposit uint64 = 1
)

func init() {
	grc20reg.Register(Token.Getter(), "")
}

func Deposit() {
	caller := runtime.PreviousRealm().Address()
	sent := banker.OriginSend()
	amount := sent.AmountOf("ugnot")

	require(uint64(amount) >= ugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d ugnot.", amount, ugnotMinDeposit))

	checkErr(adm.Mint(caller, uint64(amount)))
}

func Withdraw(amount uint64) {
	require(amount >= wugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d wugnot.", amount, wugnotMinDeposit))

	caller := runtime.PreviousRealm().Address()
	pkgaddr := runtime.CurrentRealm().Address()
	callerBal := Token.BalanceOf(caller)
	require(amount <= callerBal, ufmt.Sprintf("Insufficient balance: %d available, %d needed.", callerBal, amount))

	// send swapped ugnots to qcaller
	stdBanker := banker.NewBanker(banker.BankerTypeRealmSend)
	send := banker.Coins{{"ugnot", int64(amount)}}
	stdBanker.SendCoins(pkgaddr, caller, send)
	checkErr(adm.Burn(caller, amount))
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return Token.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := chain.Address(parts[1])
		balance := Token.BalanceOf(owner)
		return ufmt.Sprintf("%d", balance)
	default:
		return "404"
	}
}

func TotalSupply() uint64 { return Token.TotalSupply() }

func BalanceOf(owner chain.Address) uint64 {
	return Token.BalanceOf(owner)
}

func Allowance(owner, spender chain.Address) uint64 {
	return Token.Allowance(owner, spender)
}

func Transfer(to chain.Address, amount uint64) {
	userTeller := Token.CallerTeller()
	checkErr(userTeller.Transfer(to, amount))
}

func Approve(spender chain.Address, amount uint64) {
	userTeller := Token.CallerTeller()
	checkErr(userTeller.Approve(spender, amount))
}

func TransferFrom(from, to chain.Address, amount uint64) {
	userTeller := Token.CallerTeller()
	checkErr(userTeller.TransferFrom(from, to, amount))
}

func require(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
