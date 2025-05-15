package authz

import (
	"std"
)

// CoinAuthority implements Authority by checking if the caller has sufficient
// coins for specific actions. It maintains a whitelist of actions and their
// required minimum coin amounts.
//
// Example:
//
//	minCoins := std.Coins{{"ugnot", 1000}}
//	actions := []string{"propose", "vote"}
//	auth := NewCoinAuthority(actions, minCoins)
//	if auth.Can(caller, "propose") {
//	    // caller has at least 1000ugnot
//	}
type CoinAuthority struct {
	whitelistedActions []string
	minCoins           std.Coins
}

// NewCoinAuthority creates a new CoinAuthority with the given whitelisted
// actions and minimum required coins.
func NewCoinAuthority(whitelistedActions []string, minCoins std.Coins) *CoinAuthority {
	return &CoinAuthority{
		whitelistedActions: whitelistedActions,
		minCoins:           minCoins,
	}
}

// Can checks if the caller is authorized to perform an action by verifying
// if they have sufficient coins for each required coin type.
func (a *CoinAuthority) Can(addr std.Address, action string) bool {
	// Check if action is whitelisted
	isWhitelisted := false
	for _, whitelistedAction := range a.whitelistedActions {
		if whitelistedAction == action {
			isWhitelisted = true
			break
		}
	}
	if !isWhitelisted {
		return false
	}

	// Check if caller has sufficient coins
	banker := std.NewBanker(std.BankerTypeReadonly)
	coins := banker.GetCoins(addr)
	for _, minCoin := range a.minCoins {
		balance := coins.AmountOf(minCoin.Denom)
		if balance < minCoin.Amount {
			return false
		}
	}
	return true
}
