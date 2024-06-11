package foo20

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/r/demo/users"
	"std"
	"strings"

	pusers "gno.land/p/demo/users"
)

var (
	foo      *grc20.Token
	o        *ownable.Ownable
	fauceted bool
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

// Name returns the name of the token
func Name() string {
	return foo.Name()
}

// Symbol returns the name of the token
func Symbol() string {
	return foo.Symbol()
}

// TotalSupply returns the total supply of the token
func TotalSupply() uint64 {
	return foo.TotalSupply()
}

// Decimals returns the decimals of the token
func Decimals() uint64 {
	return foo.Decimals()
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

// Fauceted returns if the faucet for foo20 is currently open
func Fauceted() bool {
	return fauceted
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

// Faucet

func Faucet() {
	if !fauceted {
		err := ufmt.Sprintf("faucet is paused for token %s ($%s)", Name(), Symbol())
		panic(err)
	}

	// Mint 1000 tokens to caller
	caller := std.PrevRealm().Addr()
	Mint(caller, 1000*Decimals())
}

// Administration

func Mint(address pusers.AddressOrName, amount uint64) {
	if err := o.CallerIsOwner(); err != nil {
		panic("only owner can mint tokens")
	}

	foo.Mint(users.Resolve(address), amount)
}

func Burn(address pusers.AddressOrName, amount uint64) {
	if err := o.CallerIsOwner(); err != nil {
		panic("only owner can burn tokens")
	}

	foo.Burn(users.Resolve(address), amount)
}

func EnableFaucet() {
	if err := o.CallerIsOwner(); err != nil {
		panic("only owner can mint tokens")
	}

	fauceted = true
}

func DisableFaucet() {
	if err := o.CallerIsOwner(); err != nil {
		panic("only owner can mint tokens")
	}

	fauceted = false
}

// Render

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return foo.RenderHome()
	case c == 2 && parts[0] == "balance": // pkgpath:balance/address_or_username
		owner := pusers.AddressOrName(parts[1])
		balance, _ := foo.BalanceOf(users.Resolve(owner))
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}
