package coinflip

import (
	"std"
	"strconv"
)

type Bet struct {
    Address string
    Result bool
    Amount int64
}

var latestBets = []Bet{}

func flip(choice int) bool {
	if choice != 0 && choice != 1 {
		panic("Invalid choice")
	}

	result := int(std.ChainHeight() % 2)

	return result == choice
}

func checkMinBet(betCoins std.Coins) {
	minBet := std.NewCoin("ugnot", 1)

	if betCoins.AmountOf("ugnot") < minBet.Amount {
    	panic("mise trop faible")
	}
}

func sendPrize(caller std.Address, amount int64) {
	banker := std.NewBanker(std.BankerTypeOriginSend)
	wonCoins := std.NewCoins(std.NewCoin("ugnot", amount))

	banker.SendCoins(std.CurrentRealm().Address(), caller, wonCoins)
}

func betLogic(flipResult bool) {
	caller := std.OriginCaller()
	betCoins := std.OriginSend()
	amount := betCoins.AmountOf("ugnot")

	checkMinBet(betCoins)

	if flipResult == true {
		amount *= 2
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

	flipResult := flip(1)
	betLogic(flipResult)
}

func GetLatestResults() string {
	out := ""
	for _, bet := range latestBets {
		out += "- **" + bet.Address[:10] + "**: "
		if bet.Result == true {
		    out += "won " + strconv.Itoa(int(bet.Amount)) + "ugnot 📈"
		} else {
		    out += "lost " + strconv.Itoa(int(bet.Amount)) + "ugnot 📉"
		}
		out += "\n"
	}
	return out;
}

func GetBalance(cur realm) string {
	banker := std.NewBanker(std.BankerTypeReadonly)
	coins := banker.GetCoins(std.CurrentRealm().Address())

	std.Emit("Balance", coins.String(), strconv.Itoa(int(coins.AmountOf("ugnot"))))
	return coins.String()
}
