// Package wugnot wraps ugnot into a GRC20 token, one wugnot per ugnot.
//
// The token is a gno.land/p/nt/grc20/v1 token, so its public amounts are
// uint256, and it is registered in gno.land/r/nt/grc20reg/v1. ugnot are chain
// coins (int64), so the supply can never exceed math.MaxInt64: Deposit widens
// the sent amount, and Withdraw range-checks the amount back down to int64.
package wugnot

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"chain/runtime/unsafe"
	"math"
	"strings"

	"gno.land/p/nt/grc20/v1"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/nt/grc20reg/v1"
)

var (
	Token *grc20.Token
	adm   *grc20.PrivateLedger
)

const (
	ugnotMinDeposit  int64   = 1000
	wugnotMinDeposit uint256 = 1
)

func init(cur realm) {
	// wugnot only ever creates this one token, so id 0 can't collide.
	Token, adm = grc20.NewToken("wrapped GNOT", "wugnot", 0, 0, cur)
	grc20reg.Register(cross(cur), Token, "")
}

func Deposit(cur realm) {
	// Prevent cross-realm MITM: without this, an intermediary could
	// deposit on behalf of the caller and mint wugnot to itself
	// instead of the actual sender.
	runtime.AssertOriginCall()
	caller := cur.Previous().Address()
	sent := unsafe.OriginSend()
	amount := sent.AmountOf("ugnot")

	require(int64(amount) >= ugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d ugnot.", amount, ugnotMinDeposit))

	// amount >= ugnotMinDeposit > 0, so widening to uint256 is exact.
	checkErr(adm.Mint(caller, uint256(amount)))
}

func Withdraw(cur realm, amount uint256) {
	runtime.AssertOriginCall()
	require(amount >= wugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d wugnot.", amount, wugnotMinDeposit))

	caller := cur.Previous().Address()
	pkgaddr := cur.Address()
	callerBal := Token.BalanceOf(caller)
	require(amount <= callerBal, ufmt.Sprintf("Insufficient balance: %d available, %d needed.", callerBal, amount))
	// Unreachable while the backing holds, and checked anyway so the int64
	// conversion below can never truncate: wugnot is minted only by Deposit,
	// for ugnot that has landed at this realm, and burned on every Withdraw,
	// so the supply never exceeds this realm's ugnot balance. That balance is
	// an int64 chain coin amount, so supply, and any single balance, is at most
	// math.MaxInt64, and the balance check above has already refused anything
	// larger.
	require(amount <= math.MaxInt64, ufmt.Sprintf("Withdrawal exceeds the ugnot range: %d wugnot.", amount))

	// send swapped ugnots to qcaller
	stdBanker := banker.NewBanker(banker.BankerTypeRealmSend, cur)
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

func TotalSupply() uint256 {
	return Token.TotalSupply()
}

func BalanceOf(owner address) uint256 {
	return Token.BalanceOf(owner)
}

func Allowance(owner, spender address) uint256 {
	return Token.Allowance(owner, spender)
}

func Transfer(cur realm, to address, amount uint256) {
	userTeller := adm.CallerTeller()
	checkErr(userTeller.Transfer(0, cur, to, amount))
}

func Approve(cur realm, spender address, amount uint256) {
	userTeller := adm.CallerTeller()
	checkErr(userTeller.Approve(0, cur, spender, amount))
}

func TransferFrom(cur realm, from, to address, amount uint256) {
	userTeller := adm.CallerTeller()
	checkErr(userTeller.TransferFrom(0, cur, from, to, amount))
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
