package wugnot

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/ufmt"
	"gno.land/r/demo/defi/grc20reg"
)

var Token, adm = grc20.NewToken("wrapped GNOT", "wugnot", 0)

const (
	ugnotMinDeposit  int64 = 1000
	wugnotMinDeposit int64 = 1
)

func init() {
	grc20reg.Register(cross, Token, "")
}

func Deposit(cur realm) {
	caller := runtime.PreviousRealm().Address()
	sent := banker.OriginSend()
	amount := sent.AmountOf("ugnot")

	require(int64(amount) >= ugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d ugnot.", amount, ugnotMinDeposit))

	checkErr(adm.Mint(caller, int64(amount)))
}

func Withdraw(cur realm, amount int64) {
	require(amount >= wugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d wugnot.", amount, wugnotMinDeposit))

	caller := runtime.PreviousRealm().Address()
	pkgaddr := runtime.CurrentRealm().Address()
	callerBal := Token.BalanceOf(caller)
	require(amount <= callerBal, ufmt.Sprintf("Insufficient balance: %d available, %d needed.", callerBal, amount))

	// send swapped ugnots to qcaller
	stdBanker := banker.NewBanker(banker.BankerTypeRealmSend)
	send := chain.Coins{{"ugnot", int64(amount)}}
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
		owner := address(parts[1])
		balance := Token.BalanceOf(owner)
		return ufmt.Sprintf("%d", balance)
	default:
		return "404"
	}
}

func TotalSupply() int64 {
	return Token.TotalSupply()
}

func BalanceOf(owner address) int64 {
	return Token.BalanceOf(owner)
}

func Allowance(owner, spender address) int64 {
	return Token.Allowance(owner, spender)
}

func Transfer(cur realm, to address, amount int64) {
	userTeller := Token.CallerTeller()
	checkErr(userTeller.Transfer(to, amount))
}

func Approve(cur realm, spender address, amount int64) {
	userTeller := Token.CallerTeller()
	checkErr(userTeller.Approve(spender, amount))
}

func TransferFrom(cur realm, from, to address, amount int64) {
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
