package limitedbanker

import (
	"chain"
	"chain/banker"

	"gno.land/p/nt/ufmt"
)

// LimitedBanker is a banker that limits the total spendable coins.
// Only affects the [SendCoins] method, all others simply call the underlying
// banker's methods.
type LimitedBanker struct {
	b          banker.Banker
	spendLimit chain.Coins
}

func New(b banker.Banker, spendLimit chain.Coins) *LimitedBanker {
	if len(spendLimit) == 0 {
		spendLimit = chain.NewCoins(chain.NewCoin("ugnot", 0))
	}
	return &LimitedBanker{b: b, spendLimit: spendLimit}
}

func (lb LimitedBanker) SpendableCoins() chain.Coins {
	return lb.spendLimit
}

// Implement banker.Banker interface.
func (lb *LimitedBanker) GetCoins(addr address) chain.Coins {
	return lb.b.GetCoins(addr)
}

// Implement banker.Banker interface.
// Panics if spendLimit is exceeded.
func (lb *LimitedBanker) SendCoins(from, to address, coins chain.Coins) {
	for _, coin := range coins {
		if lb.spendLimit.AmountOf(coin.Denom) < coin.Amount {
			panic(ufmt.Sprintf("spendLimit %s exceeded for coins %s", lb.spendLimit, coins))
		}
		// Decrease limit
		for i := range lb.spendLimit {
			if coin.Denom == lb.spendLimit[i].Denom {
				lb.spendLimit[i] = lb.spendLimit[i].Sub(coin)
			}
		}
	}
	lb.b.SendCoins(from, to, coins)
}

// Implement banker.Banker interface.
func (lb *LimitedBanker) TotalCoin(denom string) int64 {
	return lb.b.TotalCoin(denom)
}

// Implement banker.Banker interface.
func (lb *LimitedBanker) IssueCoin(addr address, denom string, amount int64) {
	lb.b.IssueCoin(addr, denom, amount)
}

// Implement banker.Banker interface.
func (lb *LimitedBanker) RemoveCoin(addr address, denom string, amount int64) {
	lb.b.RemoveCoin(addr, denom, amount)
}
