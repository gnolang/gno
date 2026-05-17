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
	betLogic(cross2(cur), flipResult)
}

func Tails(cur realm) {
	runtime.AssertOriginCall()

	flipResult := flip(0)
	betLogic(cross2(cur), flipResult)
}

func GetResults(cur realm) []Bet {
	return bets
}

func GetRealmBalance(cur realm) int64 {
	banker_ := banker.NewReadonlyBanker()
	coins := banker_.GetCoins(cur.Address())
	balance := coins.AmountOf(denom)

	return balance
}

func Withdrawal(cur realm, amount int64) {
	if !config.IsAuthorized(cur.Previous().Address()) {
		panic(config.ErrUnauthorized)
	}

	runtime.AssertOriginCall()

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend, cur)

	if GetRealmBalance(cur) < amount {
		panic("insufficient balance")
	}

	banker_.SendCoins(cur.Address(), runtime.OriginCaller(), chain.NewCoins(chain.NewCoin(denom, amount)))
}

func flip(choice int) bool {
	seed1 := uint64(entropy.New().Seed())
	seed2 := uint64(entropy.New().Seed())

	r := rand.New(rand.NewPCG(seed1, seed2))
	result := r.IntN(2)

	return result == choice
}

func getBet(cur realm) chain.Coin {
	bet := banker.OriginSend()
	if len(bet) != 1 || bet[0].Denom != denom {
		panic("bet is invalid")
	}

	amount := bet[0].Amount

	if amount < minBetAmount {
		panic("Bet too low")
	}
	if amount*2 > GetRealmBalance(cross2(cur)) {
		panic("Bet too high")
	}

	return bet[0]
}

func sendPrize(cur realm, caller address, amount int64) {
	banker_ := banker.NewBanker(banker.BankerTypeRealmSend, cur)
	wonCoins := chain.NewCoins(chain.NewCoin(denom, amount))

	banker_.SendCoins(cur.Address(), caller, wonCoins)
}

func betLogic(cur realm, flipResult bool) {
	caller := runtime.OriginCaller()
	bet := getBet(cross2(cur))

	if flipResult == true {
		sendPrize(cross2(cur), caller, bet.Amount*2)
		chain.Emit("BetResult", "Result", "win")
	} else {
		chain.Emit("BetResult", "Result", "lose")
	}

	bets = append(bets, Bet{Address: caller, Result: flipResult, Amount: bet})
	latestBets.Append(Bet{Address: caller, Result: flipResult, Amount: bet})
}
