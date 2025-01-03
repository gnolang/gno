package payrolls

import (
	"errors"
	"std"
	"strings"

	"gno.land/r/demo/grc20reg"
)

// Coins transforms coins into payrolls coins, prefixing them with their type. Examples: "ugnot" -> "/native/ugnot", "gno.land/r/demo/foo20" -> "/grc20/gno.land/r/demo/foo20"
func Coins(native std.Coins, grc20 std.Coins) (std.Coins, error) {
	out := make(std.Coins, len(native)+len(grc20))
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

func addCoins(a std.Coins, b std.Coins) std.Coins {
	out := make(std.Coins, len(a))
	copy(out, a)
	for _, coin := range b {
		out = addCoinAmount(out, coin)
	}
	return out
}

func subCoins(a std.Coins, b std.Coins) std.Coins {
	out := make(std.Coins, len(a))
	copy(out, a)
	for _, coin := range b {
		out = addCoinAmount(out, std.NewCoin(coin.Denom, -coin.Amount))
	}
	return out
}

func coinsHasPositive(coins std.Coins) bool {
	for _, coin := range coins {
		if coin.Amount > 0 {
			return true
		}
	}
	return false
}

func clampCoins(a std.Coins, max std.Coins) std.Coins {
	out := std.Coins{}
	for _, coin := range a {
		maxAmount := max.AmountOf(coin.Denom)
		if coin.Amount > maxAmount {
			out = append(out, std.NewCoin(coin.Denom, maxAmount))
			continue
		}
		out = append(out, coin)
	}
	return out
}

func addCoinAmount(coins std.Coins, value std.Coin) std.Coins {
	for i, coin := range coins {
		if coin.Denom != value.Denom {
			continue
		}

		out := make(std.Coins, len(coins))
		copy(out, coins)
		out[i].Amount += value.Amount
		return out
	}

	return append(coins, value)
}

func coinsEmpty(coins std.Coins) bool {
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

func sendCoins(dst std.Address, coins std.Coins) {
	if len(coins) == 0 {
		return
	}

	natives := std.Coins{}
	grc20s := std.Coins{}

	for _, coin := range coins {
		if coin.Amount == 0 {
			continue
		}
		if coin.Amount < 0 {
			panic(errors.New("negative send amount"))
		}

		var (
			target *std.Coins
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
		*target = addCoins(*target, std.NewCoins(std.NewCoin(denom, coin.Amount)))
	}

	if len(natives) != 0 {
		banker := std.GetBanker(std.BankerTypeRealmSend)
		banker.SendCoins(std.CurrentRealm().Addr(), dst, natives)
	}

	for _, coin := range grc20s {
		err := grc20reg.MustGet(coin.Denom)().CallerTeller().Transfer(dst, uint64(coin.Amount))
		if err != nil {
			panic(err)
		}
	}
}
