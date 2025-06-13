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

var instances avl.Tree // symbol -> *instance

type instance struct {
	token  *grc20.Token
	ledger *grc20.PrivateLedger
	admin  *ownable.Ownable
	faucet int64 // per-request amount. disabled if 0.
}

func New(cur realm, name, symbol string, decimals int, initialMint, faucet int64) {
	caller := std.PreviousRealm().Address()
	NewWithAdmin(cur, name, symbol, decimals, initialMint, faucet, caller)
}

func NewWithAdmin(cur realm, name, symbol string, decimals int, initialMint, faucet int64, admin std.Address) {
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

	grc20reg.Register(cross, token, symbol)
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

func TotalSupply(symbol string) int64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().TotalSupply()
}

func IsKnownAccount(symbol string, owner std.Address) bool {
	inst := mustGetInstance(symbol)
	return inst.token.IsKnownAccount(owner)
}

func BalanceOf(symbol string, owner std.Address) int64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().BalanceOf(owner)
}

func Allowance(symbol string, owner, spender std.Address) int64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().Allowance(owner, spender)
}

func Transfer(cur realm, symbol string, to std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.Transfer(to, amount))
}

func Approve(cur realm, symbol string, spender std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.Approve(spender, amount))
}

func TransferFrom(cur realm, symbol string, from, to std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.TransferFrom(from, to, amount))
}

// faucet.
func Faucet(cur realm, symbol string) {
	inst := mustGetInstance(symbol)
	if inst.faucet == 0 {
		panic("faucet disabled for this token")
	}
	// FIXME: add limits?
	// FIXME: add payment in gnot?
	caller := std.PreviousRealm().Address()
	checkErr(inst.ledger.Mint(caller, inst.faucet))
}

func Mint(cur realm, symbol string, to std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	inst.admin.AssertOwnedByPrevious()
	checkErr(inst.ledger.Mint(to, amount))
}

func Burn(cur realm, symbol string, from std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	inst.admin.AssertOwnedByPrevious()
	checkErr(inst.ledger.Burn(from, amount))
}

// instance admin functionality
func DropInstanceOwnership(cur realm, symbol string) {
	inst := mustGetInstance(symbol)
	checkErr(inst.admin.DropOwnershipByCurrent())
}

func TransferInstanceOwnership(cur realm, symbol string, newOwner std.Address) {
	inst := mustGetInstance(symbol)
	checkErr(inst.admin.TransferOwnership(newOwner))
}

func ListTokens() []*grc20.Token {
	tokens := make([]*grc20.Token, 0, instances.Size())

	instances.Iterate("", "", func(_ string, instI any) bool {
		inst := instI.(*instance)
		tokens = append(tokens, inst.token)
		return false
	})

	return tokens
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
