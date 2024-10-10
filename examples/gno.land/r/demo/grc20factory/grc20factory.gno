package foo20

import (
	"std"
	"strconv"
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

func SetAdmin(symbol string, newAdmin std.Address) {
	inst := mustGetInstance(symbol)
	inst.admin.AssertCallerIsOwner()
	inst.admin.TransferOwnership(newAdmin)
}

type instance struct {
	banker *grc20.Banker
	admin  *ownable.Ownable
	faucet uint64 // per-request amount. disabled if 0.
}

func (inst instance) Token() grc20.Token { return inst.banker.Token() }

func TotalInstances() uint64 {
	return uint64(instances.Size())
}

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
	// Split the path by '/' to determine the number of parts (e.g., "tokens/0/10").
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	// Case 1: If the path is empty, return a list of tokens with their admins (first 10 by default).
	case path == "":
		return listTokensAndAdmins(0, 10)
	// Case 2: If the path contains only the token symbol, show token-specific home information.
	// Example path: "FOO"
	case c == 1:
		symbol := parts[0]
		inst := mustGetInstance(symbol)
		return inst.banker.RenderHome()
	// Case 3: Retrieve the balance of a token for a specific address.
	// Example path: "FOO/balance/<ownerAddress>"
	case c == 3 && parts[1] == "balance":
		symbol := parts[0]
		inst := mustGetInstance(symbol)
		owner := std.Address(parts[2])
		balance := inst.Token().BalanceOf(owner)
		return ufmt.Sprintf("%d", balance)
	// Case 4: Pagination for listing tokens.
	// Example path: "tokens/<page>/<limit>"
	case c == 3 && parts[0] == "tokens":
		page, err := strconv.Atoi(parts[1])
		if err != nil {
			return "Page is not a number\n"
		}
		limit, err := strconv.Atoi(parts[2])
		if err != nil {
			return "Limit is not a number\n"
		}
		return listTokensAndAdmins(page, limit)
	// Default case: Return "404" for unrecognized paths.
	// Example path: "unknown/path"
	default:
		return "404\n"
	}
}

// listTokensAndAdmins returns a paginated list of tokens and their admins
func listTokensAndAdmins(page, limit int) string {
	if page < 0 || limit <= 0 {
		return "Invalid page or limit\n"
	}

	var result strings.Builder
	start := page * limit
	end := start + limit

	count := 0
	instances.Iterate("", "", func(key string, value interface{}) bool {
		if count >= end {
			return true
		}

		if count >= start {
			inst := value.(*instance)
			result.WriteString(ufmt.Sprintf(
				"Name: %s, Symbol: %s, Decimals: %d, Admin: %s, Total Supply: %d\n",
				inst.Token().GetName(),
				key,
				inst.Token().GetDecimals(),
				inst.admin.Owner().String(),
				inst.Token().TotalSupply(),
			))
		}
		count++
		return false
	})

	if result.Len() == 0 {
		return "No tokens found\n"
	}

	return result.String()
}

func mustGetInstance(symbol string) *instance {
	t, exists := instances.Get(symbol)
	if !exists {
		panic(ufmt.Sprintf("token with symbol %s does not exist", symbol))
	}
	return t.(*instance)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
