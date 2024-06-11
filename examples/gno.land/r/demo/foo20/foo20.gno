package foo20

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"

	pusers "gno.land/p/demo/users"
)

var (
	foo grc20.Token
	o   *ownable.Ownable
)

func init() {
	// Instantiate ownable
	o = ownable.New()
	o.TransferOwnership("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

	// Instantiate token
	foo = grc20.NewGRC20Token("Foo", "FOO", 4)

	// Mint 1M tokens to g125em6arxsnj49vx35f0n0z34putv5ty3376fg5
	foo.Mint(o.Owner(), 1_000_000*10000)
}

// Getters

// TotalSupply returns the total supply of the token
func TotalSupply() uint64 {
	return foo.TotalSupply()
}

// BalanceOf returns the balance of a specific username or address
func BalanceOf(owner pusers.AddressOrName) uint64 {
	balance, err := foo.BalanceOf(users.Resolve(owner))
	if err != nil {
		panic(err)
	}
	return balance
}

// Allowance returns the allowance of a spender for the owner's tokens
func Allowance(owner, spender pusers.AddressOrName) uint64 {
	allowance, err := foo.Allowance(users.Resolve(owner), users.Resolve(spender))
	if err != nil {
		panic(err)
	}
	return allowance
}

// Setters

// Transfer transfers `value` amount of tokens to address `to`, and MUST fire the Transfer event
// The function SHOULD throw if the message caller’s account balance does not have enough tokens to spend
func Transfer(to pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	foo.Transfer(caller, users.Resolve(to), amount)
}

// Approve allows `spender` to withdraw from your account multiple times, up to the `value` amount
// If this function is called again it overwrites the current allowance with value
func Approve(spender pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	foo.Approve(caller, users.Resolve(spender), amount)
}

// TransferFrom transfers `value` amount of tokens from address `from` to address `to`, and MUST fire the Transfer event
func TransferFrom(from, to pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	foo.TransferFrom(caller, users.Resolve(from), users.Resolve(to), amount)
}

// faucet.

func Faucet() {
	// FIXME: add limits?
	// FIXME: add payment in gnot?
	caller := std.PrevRealm().Addr()
	Mint(caller, 1000*10000) // 1k
}

// administration.

func Mint(address pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	foo.Mint(users.Resolve(address), amount)
}

func Burn(address pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	foo.Burn(users.Resolve(address), amount)
}

// render.
//

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return foo.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := pusers.AddressOrName(parts[1])
		balance, _ := foo.BalanceOf(users.Resolve(owner))
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}
