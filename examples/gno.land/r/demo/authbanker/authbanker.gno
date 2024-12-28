// Package authbanker provide an example auth-consumer allowing to manipulate coins with auth tokens.
//
// It is quite limited, it only supports ugnots and only EOAs can fund accounts.
//
// XXX: support grc20 to allow for realm-realm interactions
package authbanker

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/authreg"
)

const denom = "ugnot"

var vaults = make(map[string]int64)

// GetCoins returns the amount of coins in the vault identified by `entityID`
func GetCoins(entityID string) int64 {
	return vaults[entityID]
}

// SendCoins sends `amount` coins from the vault identified by `atok` to the account identified by `to`.
//
// `to` can be an entity ID or an address.
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

	if isEntityID(to) {
		vaults[to] += amount
	} else {
		realmBanker := std.GetBanker(std.BankerTypeRealmSend)
		from := std.CurrentRealm().Addr()
		coins := std.Coins{std.NewCoin(denom, amount)}
		realmBanker.SendCoins(from, std.Address(to), coins)
	}
}

// FundVault funds the vault identified by `entityID` with the `OrigSend` coins.
// It panics if it's not an `OriginCall`.
func FundVault(entityID string) {
	// XXX: maybe replace the following with `authreg.Validate(to)`
	if !isEntityID(entityID) {
		panic("invalid destination")
	}

	std.AssertOriginCall()

	sentCoins := std.GetOrigSend()
	for _, coin := range sentCoins {
		if coin.Denom != denom {
			panic(ufmt.Errorf("only %q supported", denom))
		}
		if coin.Amount <= 0 {
			panic(errors.New("send amount must be > 0"))
		}
		vaults[entityID] += coin.Amount
	}
}

func isEntityID(str string) bool {
	return strings.HasPrefix(str, "/")
}
