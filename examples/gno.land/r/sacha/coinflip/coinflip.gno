package coinflip

import (
	"std"
	"math/rand"

	"gno.land/p/demo/entropy"
	"gno.land/r/sacha/config"
	"gno.land/p/moul/fifo"
)

const denom = "ugnot"
const minBetAmount = 1

var (
	bets = []Bet{}
	latestBets = fifo.New(8)
)

func Heads(cur realm) {
	std.AssertOriginCall()

	flipResult := flip(1)
	betLogic(flipResult)
}

func Tails(cur realm) {
	std.AssertOriginCall()

	flipResult := flip(0)
	betLogic(flipResult)
}

func GetResults(cur realm) []Bet {
	return bets
}

func GetRealmBalance(cur realm) int64 {
	banker := std.NewBanker(std.BankerTypeReadonly)
	coins := banker.GetCoins(std.CurrentRealm().Address())
	balance := coins.AmountOf(denom)

	return balance
}

func Withdrawal(cur realm, amount int64) {
	if !config.IsAuthorized(std.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	std.AssertOriginCall()

	banker := std.NewBanker(std.BankerTypeRealmSend)

	if GetRealmBalance(cur) < amount {
		panic("insufficient balance")
	}

	banker.SendCoins(std.CurrentRealm().Address(), std.OriginCaller(), std.NewCoins(std.NewCoin(denom, amount)))
}

func flip(choice int) bool {
	seed1 := uint64(entropy.New().Seed())
	seed2 := uint64(entropy.New().Seed())

	r := rand.New(rand.NewPCG(seed1, seed2))
	result := r.IntN(2)

	return result == choice
}

func getBet() std.Coin {
	bet := std.OriginSend()
	if len(bet) != 1 || bet[0].Denom != denom {
		panic("bet is invalid")
	}

	amount := bet[0].Amount

	if amount < minBetAmount {
		panic("Bet too low")
	}
	if amount * 2 > GetRealmBalance(cross) {
		panic("Bet too high")
	}

	return bet[0]
}

func sendPrize(caller std.Address, amount int64) {
	banker := std.NewBanker(std.BankerTypeRealmSend)
	wonCoins := std.NewCoins(std.NewCoin(denom, amount * 2))

	banker.SendCoins(std.CurrentRealm().Address(), caller, wonCoins)
}

func betLogic(flipResult bool) {
	caller := std.OriginCaller()
	bet := getBet()

	if flipResult == true {
		sendPrize(caller, bet.Amount)
		std.Emit("BetResult", "Result", "win")
	} else {
		std.Emit("BetResult", "Result", "lose")
	}

	bets = append(bets, Bet{Address: caller, Result: flipResult, Amount: bet})
	latestBets.Append(Bet{Address: caller, Result: flipResult, Amount: bet})
}
