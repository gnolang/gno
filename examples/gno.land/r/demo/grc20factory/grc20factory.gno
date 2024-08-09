package foo20

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

var instances avl.Tree // symbol -> instance

func New(name, symbol string, decimals uint, initialMint, faucet uint64) {
	caller := std.PrevRealm().Addr()
	NewWithAdmin(name, symbol, decimals, initialMint, faucet, caller)
}

func NewWithAdmin(name, symbol string, decimals uint, initialMint, faucet uint64, admin std.Address) {
	exists := instances.Has(symbol)
	if exists {
		panic("token already exists")
	}

	banker := grc20.NewBanker(name, symbol, decimals)
	if initialMint > 0 {
		banker.Mint(admin, initialMint)
	}

	inst := instance{
		banker: banker,
		admin:  ownable.NewWithAddress(admin),
		faucet: faucet,
	}

	instances.Set(symbol, &inst)
}

type instance struct {
	banker *grc20.Banker
	admin  ownable.Ownable
	faucet uint64 // per-request amount. disabled if 0.
}

func (inst instance) Token() grc20.Token { return inst.banker.Token() }

func TotalSupply(symbol string) uint64 {
	inst := mustGetInstance(symbol)
	return inst.Token().TotalSupply()
}

func BalanceOf(symbol string, owner std.Address) uint64 {
	inst := mustGetInstance(symbol)
	return inst.Token().BalanceOf(owner)
}

func Allowance(symbol string, owner, spender std.Address) uint64 {
	inst := mustGetInstance(symbol)
	return inst.Token().Allowance(owner, spender)
}

func Transfer(symbol string, to std.Address, amount uint64) {
	inst := mustGetInstance(symbol)
	checkErr(inst.Token().Transfer(to, amount))
}

func Approve(symbol string, spender std.Address, amount uint64) {
	inst := mustGetInstance(symbol)
	checkErr(inst.Token().Approve(spender, amount))
}

func TransferFrom(symbol string, from, to std.Address, amount uint64) {
	inst := mustGetInstance(symbol)
	checkErr(inst.Token().TransferFrom(from, to, amount))
}

// faucet.
func Faucet(symbol string) {
	inst := mustGetInstance(symbol)
	if inst.faucet == 0 {
		panic("faucet disabled for this token")
	}
	// FIXME: add limits?
	// FIXME: add payment in gnot?
	caller := std.PrevRealm().Addr()
	checkErr(inst.banker.Mint(caller, inst.faucet))
}

func Mint(symbol string, to std.Address, amount uint64) {
	inst := mustGetInstance(symbol)
	inst.admin.AssertCallerIsOwner()
	checkErr(inst.banker.Mint(to, amount))
}

func Burn(symbol string, from std.Address, amount uint64) {
	inst := mustGetInstance(symbol)
	inst.admin.AssertCallerIsOwner()
	checkErr(inst.banker.Burn(from, amount))
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return "TODO: list existing tokens and admins"
	case c == 1:
		symbol := parts[0]
		inst := mustGetInstance(symbol)
		return inst.banker.RenderHome()
	case c == 3 && parts[1] == "balance":
		symbol := parts[0]
		inst := mustGetInstance(symbol)
		owner := std.Address(parts[2])
		balance := inst.Token().BalanceOf(owner)
		return ufmt.Sprintf("%d", balance)
	default:
		return "404\n"
	}
}

func mustGetInstance(symbol string) *instance {
	t, exists := instances.Get(symbol)
	if !exists {
		panic("token instance does not exist")
	}
	return t.(*instance)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
