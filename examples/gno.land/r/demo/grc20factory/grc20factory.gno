package foo20

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/grc20reg"
)

var instances avl.Tree // symbol -> instance

type instance struct {
	token  *grc20.Token
	ledger *grc20.PrivateLedger
	admin  *ownable.Ownable
	faucet uint64 // per-request amount. disabled if 0.
}

func New(name, symbol string, decimals uint, initialMint, faucet uint64) {
	crossing()

	caller := std.PreviousRealm().Address()
	NewWithAdmin(name, symbol, decimals, initialMint, faucet, caller)
}

func NewWithAdmin(name, symbol string, decimals uint, initialMint, faucet uint64, admin std.Address) {
	crossing()

	exists := instances.Has(symbol)
	if exists {
		panic("token already exists")
	}

	token, ledger := grc20.NewToken(name, symbol, decimals)
	if initialMint > 0 {
		ledger.Mint(admin, initialMint)
	}

	inst := instance{
		token:  token,
		ledger: ledger,
		admin:  ownable.NewWithAddress(admin),
		faucet: faucet,
	}
	instances.Set(symbol, &inst)
	cross(grc20reg.Register)(token.Getter(), symbol)
}

func (inst instance) Token() *grc20.Token {
	return inst.token
}

func (inst instance) CallerTeller() grc20.Teller {
	return inst.token.CallerTeller()
}

func Bank(symbol string) *grc20.Token {
	inst := mustGetInstance(symbol)
	return inst.token
}

func TotalSupply(symbol string) uint64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().TotalSupply()
}

func BalanceOf(symbol string, owner std.Address) uint64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().BalanceOf(owner)
}

func Allowance(symbol string, owner, spender std.Address) uint64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().Allowance(owner, spender)
}

func Transfer(symbol string, to std.Address, amount uint64) {
	crossing()

	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.Transfer(to, amount))
}

func Approve(symbol string, spender std.Address, amount uint64) {
	crossing()

	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.Approve(spender, amount))
}

func TransferFrom(symbol string, from, to std.Address, amount uint64) {
	crossing()

	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.TransferFrom(from, to, amount))
}

// faucet.
func Faucet(symbol string) {
	crossing()

	inst := mustGetInstance(symbol)
	if inst.faucet == 0 {
		panic("faucet disabled for this token")
	}
	// FIXME: add limits?
	// FIXME: add payment in gnot?
	caller := std.PreviousRealm().Address()
	checkErr(inst.ledger.Mint(caller, inst.faucet))
}

func Mint(symbol string, to std.Address, amount uint64) {
	crossing()

	inst := mustGetInstance(symbol)
	inst.admin.AssertOwnedByPrevious()
	checkErr(inst.ledger.Mint(to, amount))
}

func Burn(symbol string, from std.Address, amount uint64) {
	crossing()

	inst := mustGetInstance(symbol)
	inst.admin.AssertOwnedByPrevious()
	checkErr(inst.ledger.Burn(from, amount))
}

// instance admin functionality
func DropInstanceOwnership(symbol string) {
	crossing()

	inst := mustGetInstance(symbol)
	checkErr(inst.admin.DropOwnershipByCurrent())
}

func TransferInstanceOwnership(symbol string, newOwner std.Address) {
	crossing()

	inst := mustGetInstance(symbol)
	checkErr(inst.admin.TransferOwnership(newOwner))
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
		return inst.token.RenderHome()
	case c == 3 && parts[1] == "balance":
		symbol := parts[0]
		inst := mustGetInstance(symbol)
		owner := std.Address(parts[2])
		balance := inst.token.CallerTeller().BalanceOf(owner)
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
		panic(err.Error())
	}
}
