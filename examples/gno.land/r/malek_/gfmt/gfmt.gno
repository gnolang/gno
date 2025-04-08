package gfmt

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	pusers "gno.land/p/demo/users"
	"gno.land/r/demo/users"
)

var (
	banker           *grc20.Banker
	admin            *ownable.Ownable
	Token            grc20.Token
	ErrInvalidUint64 = errors.New("invalid uint64, must be between 0 and 18446744073709551615")
)

func init() {
	admin = ownable.NewWithAddress("g10ahumypepd2qcrau7kahv8q78f7jcdns5tn54a") // owner's address (malek_)
	banker = grc20.NewBanker("GnoFundMe Token", "gfmt", 0)
	banker.Mint(admin.Owner(), 100_000_000_000)
	Token = banker.Token()
}

const (
	ugnotMinDeposit uint64 = 1000000 // 1 GNOT
	gfmtMinDeposit  uint64 = 1
)

func Deposit() {
	caller := std.PrevRealm().Addr()
	sent := std.GetOrigSend()
	amount := sent.AmountOf("ugnot")

	require(uint64(amount) >= ugnotMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d ugnot.", amount, ugnotMinDeposit))
	checkErr(banker.Mint(caller, uint64(amount)))
}

func Withdraw(amount uint64) {
	// check for overflow
	// if amount < 0 || amount > 18446744073709551615 {
	// 	panic("Invalid amount: Overflow uint64")
	// }

	require(amount >= gfmtMinDeposit, ufmt.Sprintf("Deposit below minimum: %d/%d gfmt.", amount, gfmtMinDeposit))

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
		return ufmt.Sprintf("%s's balance is %dugnot", owner, balance)
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

func Burn(from pusers.AddressOrName, amount uint64) {
	admin.AssertCallerIsOwner()
	fromAddr := users.Resolve(from)
	checkErr(banker.Burn(fromAddr, amount))
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
