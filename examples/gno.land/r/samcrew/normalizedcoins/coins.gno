package normalizedcoins

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"
	"strings"

	"gno.land/r/demo/defi/grc20reg"
)

// Package normalizedcoins provides utilities for handling normalized coin representations
// that combine native and GRC20 tokens with prefixed denoms.
//
// This implementation is inspired by the payrolls realm from N0izN0iz:
// https://github.com/gnolang/gno/pull/3432

// PrefixCoins transforms coins into normalized coins, prefixing their denoms with their type.
// Native coins are prefixed with "/native/" and GRC20 coins with "/grc20/".
//
// Examples:
//   - "ugnot" -> "/native/ugnot"
//   - "gno.land/r/demo/foo20" -> "/grc20/r/demo/foo20"
//
// Returns an error if no coins are provided or if any coin has a negative amount.
func PrefixCoins(native chain.Coins, grc20 chain.Coins) (chain.Coins, error) {
	if len(native) == 0 && len(grc20) == 0 {
		return chain.Coins{}, errors.New("no coins provided")
	}
	if CoinsHasNegative(native) || CoinsHasNegative(grc20) {
		return chain.Coins{}, errors.New("negative coin amounts are not allowed")
	}
	out := make(chain.Coins, len(native)+len(grc20))
	for i, coin := range native {
		out[i].Amount = coin.Amount
		out[i].Denom = "/native/" + coin.Denom
	}
	offset := len(native)
	for i, coin := range grc20 {
		j := offset + i
		out[j].Amount = coin.Amount
		denom := strings.TrimPrefix(coin.Denom, "gno.land")
		denom = strings.TrimPrefix(denom, "/")
		out[j].Denom = "/grc20/" + denom
	}
	return out, nil
}

// CoinsHasNegative returns true if any coin in the collection has a negative amount.
func CoinsHasNegative(coins chain.Coins) bool {
	for _, coin := range coins {
		if coin.Amount < 0 {
			return true
		}
	}
	return false
}

// CoinsHasPositive returns true if any coin in the collection has a positive amount.
func CoinsHasPositive(coins chain.Coins) bool {
	for _, coin := range coins {
		if coin.Amount > 0 {
			return true
		}
	}
	return false
}

// AddCoins adds two coin collections together, combining amounts for matching denoms.
func AddCoins(a chain.Coins, b chain.Coins) chain.Coins {
	out := make(chain.Coins, len(a))
	copy(out, a)
	for _, coin := range b {
		out = AddCoinAmount(out, coin)
	}
	return out
}

// SubCoins subtracts coin collection b from a, reducing amounts for matching denoms.
func SubCoins(a chain.Coins, b chain.Coins) chain.Coins {
	out := make(chain.Coins, len(a))
	copy(out, a)
	for _, coin := range b {
		out = AddCoinAmount(out, chain.NewCoin(coin.Denom, -coin.Amount))
	}
	return out
}

// MultiplyCoins multiplies all coin amounts by the given factor.
func MultiplyCoins(coins chain.Coins, factor int64) chain.Coins {
	out := make(chain.Coins, len(coins))
	for i, coin := range coins {
		out[i] = chain.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount * factor,
		}
	}
	return out
}

// LessCoinsThan returns true if any coin amount in a is less than its corresponding amount in b.
func LessCoinsThan(a chain.Coins, b chain.Coins) bool {
	for _, coinA := range a {
		found := false
		for _, coinB := range b {
			if coinA.Denom == coinB.Denom {
				if coinA.Amount < coinB.Amount {
					return true
				}
				found = true
				break
			}
		}
		if !found && coinA.Amount > 0 {
			return false
		}
	}
	return false
}

// AddCoinAmount adds a single coin amount to a coin collection.
// If the denom exists, it adds to the existing amount; otherwise, it appends the coin.
func AddCoinAmount(coins chain.Coins, value chain.Coin) chain.Coins {
	for i, coin := range coins {
		if coin.Denom != value.Denom {
			continue
		}

		out := make(chain.Coins, len(coins))
		copy(out, coins)
		out[i].Amount += value.Amount
		return out
	}
	return append(coins, value)
}

// SendCoins sends normalized coins to a destination address.
// It automatically separates native and GRC20 coins based on their prefixes
// and uses the appropriate transfer mechanism for each type.
//
// Panics if any coin has a negative amount or an invalid denom prefix.
func SendCoins(dst address, coins chain.Coins) {
	if len(coins) == 0 {
		return
	}

	natives := chain.Coins{}
	grc20s := chain.Coins{}

	for _, coin := range coins {
		if coin.Amount == 0 {
			continue
		}
		if coin.Amount < 0 {
			panic(errors.New("negative send amount"))
		}

		if strings.HasPrefix(coin.Denom, "/native/") {
			denom := strings.TrimPrefix(coin.Denom, "/native/")
			natives = AddCoinAmount(natives, chain.NewCoin(denom, coin.Amount))
		} else if strings.HasPrefix(coin.Denom, "/grc20/") {
			denom := strings.TrimPrefix(coin.Denom, "/grc20/")
			grc20s = AddCoinAmount(grc20s, chain.NewCoin(denom, coin.Amount))
		} else {
			panic(errors.New("invalid coin denom prefix: " + coin.Denom))
		}
	}

	banker_ := banker.NewBanker(banker.BankerTypeRealmIssue)
	from := runtime.CurrentRealm().Address()
	if len(natives) != 0 {
		banker_.SendCoins(from, dst, natives)
	}

	for _, coin := range grc20s {
		grc20reg.MustGet(coin.Denom).RealmTeller().Transfer(dst, coin.Amount)
	}
}
