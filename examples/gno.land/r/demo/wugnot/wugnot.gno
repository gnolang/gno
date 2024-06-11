package wugnot

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"

	"gno.land/r/demo/users"

	pusers "gno.land/p/demo/users"
)

// wugnot is the main token struct
var wugnot *grc20.Token = grc20.NewGRC20Token("wrapped GNOT", "wugnot", 0)

const (
	ugnotMinDeposit  uint64 = 1
	wugnotMinDeposit uint64 = 1
)

// wrapper.
//

// Deposit can be called to deposit ugnots, and get wugnots in return
func Deposit() {
	caller := std.PrevRealm().Addr()
	sentAmount := std.GetOrigSend().AmountOf("ugnot")

	if uint64(sentAmount) < ugnotMinDeposit {
		panic(ufmt.Sprintf("Deposit below minimum: %d/%d ugnot.", sentAmount, ugnotMinDeposit))
	}

	wugnot.Mint(caller, uint64(sentAmount))
}

// Withdraw can be called to unwrap wugnots back into ugnots
func Withdraw(amount uint64) {
	if amount < wugnotMinDeposit {
		panic(ufmt.Sprintf("Deposit below minimum: %d/%d wugnot.", amount, wugnotMinDeposit))
	}

	caller := std.PrevRealm().Addr()
	wugnotAddress := std.CurrentRealm().Addr()

	callerBal := wugnot.BalanceOf(caller)
	if callerBal < amount {
		panic(ufmt.Sprintf("Insufficient balance: %d available, %d needed.", callerBal, amount))
	}

	// First burn wugnot balance
	wugnot.Burn(caller, amount)

	// unwrap wugnot and send ugnot to caller
	banker := std.GetBanker(std.BankerTypeRealmSend)
	send := std.NewCoins(std.NewCoin("ugnot", int64(amount)))
	banker.SendCoins(wugnotAddress, caller, send)
}

// Rendering

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return wugnot.RenderHome()
	case c == 2 && parts[0] == "balance": // pkgpath:balance/address_or_username
		owner := pusers.AddressOrName(parts[1])
		return ufmt.Sprintf("%d\n", wugnot.BalanceOf(users.Resolve(owner)))
	default:
		return "404\n"
	}
}

// XXX: if we could call WUGNOT.XXX instead of XXX from gnokey, then, all the following lines would not be needed.

// direct getters.
// XXX: remove them in favor of q_call wugnot.XXX

// Name returns the name of the token
func Name() string {
	return wugnot.Name()
}

// Symbol returns the name of the token
func Symbol() string {
	return wugnot.Symbol()
}

// TotalSupply returns the total supply of the token
func TotalSupply() uint64 {
	return wugnot.TotalSupply()
}

// Decimals returns the decimals of the token
func Decimals() uint8 {
	return wugnot.Decimals()
}

// BalanceOf returns the balance of a specific username or address
func BalanceOf(owner pusers.AddressOrName) uint64 {
	return wugnot.BalanceOf(users.Resolve(owner))
}

// Allowance returns the allowance of a spender for the owner's tokens
func Allowance(owner, spender pusers.AddressOrName) uint64 {
	return wugnot.Allowance(users.Resolve(owner), users.Resolve(spender))
}

// Setters

// Transfer transfers `value` amount of tokens to address `to`, and MUST fire the Transfer event
// The function SHOULD throw if the message caller’s account balance does not have enough tokens to spend
func Transfer(to pusers.AddressOrName, amount uint64) {
	wugnot.Transfer(users.Resolve(to), amount)
}

// Approve allows `spender` to withdraw from your account multiple times, up to the `value` amount
// If this function is called again it overwrites the current allowance with value
func Approve(spender pusers.AddressOrName, amount uint64) {
	wugnot.Approve(users.Resolve(spender), amount)
}

// TransferFrom transfers `value` amount of tokens from address `from` to address `to`, and MUST fire the Transfer event
func TransferFrom(from, to pusers.AddressOrName, amount uint64) {
	wugnot.TransferFrom(users.Resolve(from), users.Resolve(to), amount)
}
