package authbanker

import (
	"std"
	"strings"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/authreg"
)

// This example is there mostly to demonstrate auth usage, it is quite limited
// only EOAs can fund accounts

const denom = "ugnot"

var vaults = make(map[string]int64)

func GetCoins(account string) int64 {
	return vaults[account]
}

func SendCoins(atok auth.Token, to string, amount int64) {
	if amount < 1 {
		panic("sent amount must be >= 0")
	}

	from := authreg.Authenticate(atok)

	if from == to {
		panic("cannot send to self")
	}
	if vaultAmount := vaults[from]; amount > vaultAmount {
		panic("not enough in account")
	}

	vaults[from] -= amount

	if isEntityId(to) {
		vaults[to] += amount
	} else {
		realmBanker := std.GetBanker(std.BankerTypeRealmSend)
		from := std.CurrentRealm().Addr()
		coins := std.Coins{std.NewCoin(denom, amount)}
		realmBanker.SendCoins(from, std.Address(to), coins)
	}
}

func FundVault(to string) {
	// XXX: maybe replace the following with `authreg.Validate(to)`
	if !isEntityId(to) {
		panic("invalid destination")
	}

	std.AssertOriginCall()

	sentCoins := std.GetOrigSend()
	for _, coin := range sentCoins {
		if coin.Denom != denom {
			panic(ufmt.Errorf("only %q supported", denom))
		}
		vaults[to] += coin.Amount
	}
}

func isEntityId(str string) bool {
	return strings.HasPrefix(str, "/")
}
