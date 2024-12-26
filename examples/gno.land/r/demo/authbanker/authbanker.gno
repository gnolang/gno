package authbanker

import (
	"std"
	"strings"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/authreg"
)

// this example is there mostly to demonstrate auth usage, it is quite limited
// only EOAs can fund accounts

const denom = "ugnot"

var vaults = make(map[string]int64)

func GetCoins(account string) int64 {
	return vaults[account]
}

func SendCoins(atok auth.Token, to string, amount int64) {
	account := authreg.Authenticate(atok)

	if amount < 1 {
		panic(ufmt.Errorf("sent amount of %q must be >= 0", denom))
	}
	if vaultAmount := vaults[account]; amount > vaultAmount {
		panic(ufmt.Errorf("not enough %q in account, wanted %d, got %d", denom, amount, vaultAmount))
	}
	if account == to {
		panic("cannot send to self")
	}

	vaults[account] -= amount

	if strings.HasPrefix(to, "/") {
		vaults[to] += amount
	} else {
		realmBanker := std.GetBanker(std.BankerTypeRealmSend)
		from := std.CurrentRealm().Addr()
		coins := std.Coins{std.NewCoin(denom, amount)}
		realmBanker.SendCoins(from, std.Address(to), coins)
	}
}

func FundVault(to string) {
	if !strings.HasPrefix(to, "/") {
		panic("invalid destination")
	}
	// XXX: maybe `authreg.Validate(to)`
	std.AssertOriginCall()
	sentCoins := std.GetOrigSend()
	for _, coin := range sentCoins {
		if coin.Denom != denom {
			panic(ufmt.Errorf("only %q supported", denom))
		}
		vaults[to] += coin.Amount
	}
}

func TotalCoin(denom string) int64 {
	return std.GetBanker(std.BankerTypeRealmSend).TotalCoin(denom)
}
