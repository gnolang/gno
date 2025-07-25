package coinflip

import (
	"std"
	"strconv"
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

	result := int(std.ChainHeight() % 2)

	return result == choice
}

func checkMinBet(betCoins std.Coins) {
	minBet := std.NewCoin(denom, 1)

	if betCoins.AmountOf(denom) < minBet.Amount {
    	panic("mise trop faible")
	}
}

func sendPrize(caller std.Address, amount int64) {
	banker := std.NewBanker(std.BankerTypeRealmSend)
	wonCoins := std.NewCoins(std.NewCoin(denom, amount))

	banker.SendCoins(std.CurrentRealm().Address(), caller, wonCoins)
}

func betLogic(flipResult bool) {
	caller := std.OriginCaller()
	betCoins := std.OriginSend()
	amount := betCoins.AmountOf(denom)

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
		out += "- **" + bet.Address[:30] + "**: "
		if bet.Result == true {
		    out += "won " + strconv.Itoa(int(bet.Amount)) + denom + " 📈"
		} else {
		    out += "lost " + strconv.Itoa(int(bet.Amount)) + denom + " 📉"
		}
		out += "\n"
	}
	return out;
}

func GetBalance(cur realm) int64 {
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
	std.AssertOriginCall()

	banker := std.NewBanker(std.BankerTypeRealmSend)
	realmBalance := banker.GetCoins(std.CurrentRealm().Address()).AmountOf(denom);

	if realmBalance < amount {
		panic("insufficient balance")
	}

	banker.SendCoins(std.CurrentRealm().Address(), std.OriginCaller(), std.NewCoins(std.NewCoin(denom, amount)))
}
