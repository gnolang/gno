package wugnot

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

var (
	// wugnot is the admin token, able to mint and burn.
	wugnot *grc20.AdminToken = grc20.NewAdminToken("wrapped GNOT", "wugnot", 0)
	// WUGNOT is the banker usable by users directly.
	WUGNOT = wugnot.GRC20()
)

const (
	ugnotMinDeposit  uint64 = 1000
	wugnotMinDeposit uint64 = 1
)

// wrapper.
//

func Deposit() {
	caller := std.PrevRealm().Addr()
	sent := std.GetOrigSend()
	amount := sent.AmountOf("ugnot")

	if uint64(amount) < ugnotMinDeposit {
		panic(ufmt.Sprintf("Deposit below minimum: %d/%d ugnot.", amount, ugnotMinDeposit))
	}
	wugnot.Mint(caller, uint64(amount))
}

func Withdraw(amount uint64) {
	if amount < wugnotMinDeposit {
		panic(ufmt.Sprintf("Deposit below minimum: %d/%d wugnot.", amount, wugnotMinDeposit))
	}

	caller := std.PrevRealm().Addr()
	pkgaddr := std.CurrentRealm().Addr()

	callerBal, _ := wugnot.BalanceOf(caller)
	if callerBal < amount {
		panic(ufmt.Sprintf("Insufficient balance: %d available, %d needed.", callerBal, amount))
	}

	// send swapped ugnots to caller
	banker := std.GetBanker(std.BankerTypeRealmSend)
	send := std.Coins{{"ugnot", int64(amount)}}
	banker.SendCoins(pkgaddr, caller, send)
	wugnot.Burn(caller, amount)
}

// render.
//

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return wugnot.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := std.Address(parts[1])
		balance, _ := wugnot.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

// XXX: if we could call WUGNOT.XXX instead of XXX from gnokey, then, all the following lines would not be needed.

// direct getters.
// XXX: remove them in favor of q_call wugnot.XXX

func TotalSupply() uint64 {
	return wugnot.TotalSupply()
}

func BalanceOf(owner std.Address) uint64 {
	balance, err := wugnot.BalanceOf(owner)
	if err != nil {
		panic(err)
	}
	return balance
}

func Allowance(owner, spender std.Address) uint64 {
	allowance, err := wugnot.Allowance(owner, spender)
	if err != nil {
		panic(err)
	}
	return allowance
}

// setters.
//

func Transfer(to std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := wugnot.Transfer(caller, to, amount)
	if err != nil {
		panic(err)
	}
}

func Approve(spender std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := wugnot.Approve(caller, spender, amount)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := wugnot.TransferFrom(caller, from, to, amount)
	if err != nil {
		panic(err)
	}
}
