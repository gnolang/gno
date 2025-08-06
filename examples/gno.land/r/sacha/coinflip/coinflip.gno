package coinflip

import (
	"std"
	"strconv"
	"math/rand"
	
	"gno.land/p/demo/entropy"
	"gno.land/r/sacha/config"
)

type Bet struct {
    Address string
    Result bool
    Amount int64
}

var (
	latestBets = []Bet{}
	denom string
)


func init() {
	denom = "ugnot"
}

func flip(choice int) bool {
	if choice != 0 && choice != 1 {
		panic("Invalid choice")
	}

	seed1 := uint64(entropy.New().Seed())
	seed2 := uint64(entropy.New().Seed())

	r    := rand.New(rand.NewPCG(seed1, seed2))
	result := r.IntN(2)

	return result == choice
}

func checkBet(betCoins std.Coins) {
	minBet := std.NewCoin(denom, 1)

	if betCoins.AmountOf(denom) < minBet.Amount {
		panic("Bet too low")
	}
	if betCoins.AmountOf(denom) * 2 > GetRealmBalance(cross) {
		panic("Bet too high")
	}
}

func sendPrize(caller std.Address, amount int64) {
	banker := std.NewBanker(std.BankerTypeRealmSend)
	wonCoins := std.NewCoins(std.NewCoin(denom, amount * 2))

	banker.SendCoins(std.CurrentRealm().Address(), caller, wonCoins)
}

func betLogic(flipResult bool) {
	caller := std.OriginCaller()
	betCoins := std.OriginSend()
	amount := betCoins.AmountOf(denom)

	checkBet(betCoins)

	if flipResult == true {
		sendPrize(caller, amount)
		std.Emit("BetResult", "Result", "win")
	} else {
		std.Emit("BetResult", "Result", "lose")
	}

	latestBets = append([]Bet{{Address: caller.String(), Result: flipResult, Amount: amount}}, latestBets...)
}

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

func UpdateCoinUsed(cur realm, newCoinDenom string) {
	if !config.IsAuthorized(std.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	denom = newCoinDenom
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
