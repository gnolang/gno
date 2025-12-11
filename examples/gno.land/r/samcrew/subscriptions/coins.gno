package subscriptions

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"
	"strings"

	"gno.land/r/demo/defi/grc20reg"
)

// XXX: this part is 100% from payrolls realm from N0izN0iz <3
// Payrolls realm: https://github.com/gnolang/gno/pull/3432

// Coins transforms coins into subscriptions coins, prefixing their denom them with their type. Examples: "ugnot" -> "/native/ugnot", "gno.land/r/demo/foo20" -> "/grc20/gno.land/r/demo/foo20"
func Coins(cur realm, native chain.Coins, grc20 chain.Coins) (chain.Coins, error) {
	if len(native) == 0 && len(grc20) == 0 {
		return chain.Coins{}, errors.New("no coins provided")
	}
	if coinsHasNegative(native) || coinsHasNegative(grc20) {
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
		out[j].Denom = "/grc20/" + coin.Denom
	}
	return out, nil
}

func coinsHasNegative(coins chain.Coins) bool {
	for _, coin := range coins {
		if coin.Amount < 0 {
			return true
		}
	}
	return false
}

func coinsHasPositive(coins chain.Coins) bool {
	for _, coin := range coins {
		if coin.Amount > 0 {
			return true
		}
	}
	return false
}

func addCoins(a chain.Coins, b chain.Coins) chain.Coins {
	out := make(chain.Coins, len(a))
	copy(out, a)
	for _, coin := range b {
		out = addCoinAmount(out, coin)
	}
	return out
}

func subCoins(a chain.Coins, b chain.Coins) chain.Coins {
	out := make(chain.Coins, len(a))
	copy(out, a)
	for _, coin := range b {
		out = addCoinAmount(out, chain.NewCoin(coin.Denom, -coin.Amount))
	}
	return out
}

func multiplyCoins(coins chain.Coins, factor int64) chain.Coins {
	out := make(chain.Coins, len(coins))
	for i, coin := range coins {
		out[i] = chain.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount * factor,
		}
	}
	return out
}

func lessCoinsThan(a chain.Coins, b chain.Coins) bool {
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

func addCoinAmount(coins chain.Coins, value chain.Coin) chain.Coins {
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

func sendCoins(dst address, coins chain.Coins) {
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

		var (
			target *chain.Coins
			denom  string
		)

		if strings.HasPrefix(coin.Denom, "/native/") {
			target = &natives
			denom = strings.TrimPrefix(coin.Denom, "/native/")
		} else if strings.HasPrefix(coin.Denom, "/grc20/") {
			target = &grc20s
			denom = strings.TrimPrefix(coin.Denom, "/grc20/")
		} else {
			panic(errors.New("invalid coin denom prefix: " + coin.Denom))
		}
		*target = addCoinAmount(*target, chain.NewCoin(denom, coin.Amount))
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
