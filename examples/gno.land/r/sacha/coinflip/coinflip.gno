package coinflip

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"math/rand"

	"gno.land/p/demo/entropy"
	"gno.land/p/moul/fifo"
	"gno.land/r/sacha/config"
)

const denom = "ugnot"
const minBetAmount = 1

var (
	bets       = []Bet{}
	latestBets = fifo.New(8)
)

func Heads(cur realm) {
	runtime.AssertOriginCall()

	flipResult := flip(1)
	betLogic(flipResult)
}

func Tails(cur realm) {
	runtime.AssertOriginCall()

	flipResult := flip(0)
	betLogic(flipResult)
}

func GetResults(cur realm) []Bet {
	return bets
}

func GetRealmBalance(cur realm) int64 {
	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	coins := banker_.GetCoins(runtime.CurrentRealm().Address())
	balance := coins.AmountOf(denom)

	return balance
}

func Withdrawal(cur realm, amount int64) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	runtime.AssertOriginCall()

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)

	if GetRealmBalance(cur) < amount {
		panic("insufficient balance")
	}

	banker_.SendCoins(runtime.CurrentRealm().Address(), runtime.OriginCaller(), chain.NewCoins(chain.NewCoin(denom, amount)))
}

func flip(choice int) bool {
	seed1 := uint64(entropy.New().Seed())
	seed2 := uint64(entropy.New().Seed())

	r := rand.New(rand.NewPCG(seed1, seed2))
	result := r.IntN(2)

	return result == choice
}

func getBet() chain.Coin {
	bet := banker.OriginSend()
	if len(bet) != 1 || bet[0].Denom != denom {
		panic("bet is invalid")
	}

	amount := bet[0].Amount

	if amount < minBetAmount {
		panic("Bet too low")
	}
	if amount*2 > GetRealmBalance(cross) {
		panic("Bet too high")
	}

	return bet[0]
}

func sendPrize(cur realm, caller address, amount int64) {
	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	wonCoins := chain.NewCoins(chain.NewCoin(denom, amount))

	banker_.SendCoins(runtime.CurrentRealm().Address(), caller, wonCoins)
}

func betLogic(flipResult bool) {
	caller := runtime.OriginCaller()
	bet := getBet()

	if flipResult == true {
		sendPrize(cross, caller, bet.Amount*2)
		chain.Emit("BetResult", "Result", "win")
	} else {
		chain.Emit("BetResult", "Result", "lose")
	}

	bets = append(bets, Bet{Address: caller, Result: flipResult, Amount: bet})
	latestBets.Append(Bet{Address: caller, Result: flipResult, Amount: bet})
}
