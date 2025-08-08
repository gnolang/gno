package coinflip

import (
	"std"
	"strconv"
	"gno.land/r/sacha/config"
	"math/rand"
	"gno.land/p/demo/entropy"
)

type Bet struct {
    Address string
    Result bool
    Amount int64
}

const denom = "ugnot"
const minBetAmount = 1

var latestBets = []Bet{}

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

func GetLatestResults(cur realm) [][]string {
	latestResults := [][]string{}

	for _, bet := range latestBets {
		betStr := []string{bet.Address}
		if bet.Result == true {
		    betStr = append(betStr, "won")
			betStr = append(betStr, "📈")
		} else {
		    betStr = append(betStr, "lost")
			betStr = append(betStr, "📉")
		}
		betStr = append(betStr, strconv.Itoa(int(bet.Amount)))
		latestResults = append(latestResults, betStr)
	}
	return latestResults;
}

func GetRealmBalance(cur realm) int64 {
	banker := std.NewBanker(std.BankerTypeReadonly)
	coins := banker.GetCoins(std.CurrentRealm().Address())
	balance := coins.AmountOf(denom)

	return balance
}

func CashOut(cur realm, amount int64) {
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

func checkBet(bet std.Coins) int64 {
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

	return amount
}

func sendPrize(caller std.Address, amount int64) {
	banker := std.NewBanker(std.BankerTypeRealmSend)
	wonCoins := std.NewCoins(std.NewCoin(denom, amount * 2))

	banker.SendCoins(std.CurrentRealm().Address(), caller, wonCoins)
}

func betLogic(flipResult bool) {
	caller := std.OriginCaller()
	amount := checkBet(std.OriginSend())

	if flipResult == true {
		sendPrize(caller, amount)
		std.Emit("BetResult", "Result", "win")
	} else {
		std.Emit("BetResult", "Result", "lose")
	}

	latestBets = append([]Bet{{Address: caller.String(), Result: flipResult, Amount: amount}}, latestBets...)
}
