package payrolls

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"
	"strings"

	grc20reg "gno.land/r/demo/defi/grc20reg"
)

// Coins transforms coins into payrolls coins, prefixing their denom them with their type. Examples: "ugnot" -> "/native/ugnot", "gno.land/r/demo/foo20" -> "/grc20/gno.land/r/demo/foo20"
func Coins(cur realm, native chain.Coins, grc20 chain.Coins) (chain.Coins, error) {
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

func coinsHasPositive(coins chain.Coins) bool {
	for _, coin := range coins {
		if coin.Amount > 0 {
			return true
		}
	}
	return false
}

func clampCoins(a chain.Coins, max chain.Coins) chain.Coins {
	out := chain.Coins{}
	for _, coin := range a {
		maxAmount := max.AmountOf(coin.Denom)
		if coin.Amount > maxAmount {
			out = append(out, chain.NewCoin(coin.Denom, maxAmount))
			continue
		}
		out = append(out, coin)
	}
	return out
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

func coinsEmpty(coins chain.Coins) bool {
	if len(coins) == 0 {
		return true
	}
	for _, coin := range coins {
		if coin.Amount != 0 {
			return false
		}
	}
	return true
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
			panic(errors.New("unknown coin kind"))
		}
		*target = addCoins(*target, chain.NewCoins(chain.NewCoin(denom, coin.Amount)))
	}

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	from := runtime.CurrentRealm().Address()
	if len(natives) != 0 {
		banker_.SendCoins(from, dst, natives)
	}

	for _, coin := range grc20s {
		grc20reg.MustGet(coin.Denom).RealmTeller().Transfer(dst, coin.Amount)
	}
}
