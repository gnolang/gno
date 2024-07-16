package wugnot

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	pusers "gno.land/p/demo/users"
	"gno.land/r/demo/users"
)

var (
	banker *grc20.Banker = grc20.NewBanker("wrapped GNOT", "wugnot", 0)
	Token                = banker.Token()
)

const (
	ugnotMinDeposit  uint64 = 1000
	wugnotMinDeposit uint64 = 1
)

func Deposit() {
	caller := std.PrevRealm().Addr()
	sent := std.GetOrigSend()
	amount := sent.AmountOf("ugnot")

	require(uint64(amount) >= ugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d ugnot.", amount, ugnotMinDeposit))
	checkErr(banker.Mint(caller, uint64(amount)))
}

func Withdraw(amount uint64) {
	require(amount >= wugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d wugnot.", amount, wugnotMinDeposit))

	caller := std.PrevRealm().Addr()
	pkgaddr := std.CurrentRealm().Addr()
	callerBal := Token.BalanceOf(caller)
	require(amount <= callerBal, ufmt.Sprintf("Insufficient balance: %d available, %d needed.", callerBal, amount))

	// send swapped ugnots to qcaller
	stdBanker := std.GetBanker(std.BankerTypeRealmSend)
	send := std.Coins{{"ugnot", int64(amount)}}
	stdBanker.SendCoins(pkgaddr, caller, send)
	checkErr(banker.Burn(caller, amount))
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return banker.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := std.Address(parts[1])
		balance := Token.BalanceOf(owner)
		return ufmt.Sprintf("%d", balance)
	default:
		return "404"
	}
}

func TotalSupply() uint64 { return Token.TotalSupply() }

func BalanceOf(owner pusers.AddressOrName) uint64 {
	ownerAddr := users.Resolve(owner)
	return Token.BalanceOf(ownerAddr)
}

func Allowance(owner, spender pusers.AddressOrName) uint64 {
	ownerAddr := users.Resolve(owner)
	spenderAddr := users.Resolve(spender)
	return Token.Allowance(ownerAddr, spenderAddr)
}

func Transfer(to pusers.AddressOrName, amount uint64) {
	toAddr := users.Resolve(to)
	checkErr(Token.Transfer(toAddr, amount))
}

func Approve(spender pusers.AddressOrName, amount uint64) {
	spenderAddr := users.Resolve(spender)
	checkErr(Token.Approve(spenderAddr, amount))
}

func TransferFrom(from, to pusers.AddressOrName, amount uint64) {
	fromAddr := users.Resolve(from)
	toAddr := users.Resolve(to)
	checkErr(Token.TransferFrom(fromAddr, toAddr, amount))
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
