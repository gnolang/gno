package disperse

import (
	"std"
)

// SendSingleCoin parses receivers and amounts and sends out a single coin
func SendSingleCoin(addresses []std.Address, coins std.Coins) {
	coinSent := std.GetOrigSend() // get Coins sent with call
	caller := std.GetOrigCaller() // get tx sender

	if len(coinSent) != len(coins) {
		panic(ErrArgLenAndSentLenMismatch)
	}

	if len(addresses) != len(coins) {
		panic(errNumAddrValMismatch)
	}

	for _, coin := range coins {
		if coins.AmountOf(coin.Denom) != coinSent.AmountOf(coin.Denom) {
			panic(ErrWrongAmount)
		}
	}

	// Get address of Disperse realm
	realmAddr := std.CurrentRealm().Addr()

	// Get Banker
	banker := std.GetBanker(std.BankerTypeOrigSend)

	// Send coins
	for i, _ := range addresses {
		banker.SendCoins(realmAddr, addresses[i], std.Coins{coins[i]})
	}

	// Return possible leftover coins
	for _, coin := range coinSent {
		leftoverAmt := banker.GetCoins(realmAddr).AmountOf(coin.Denom)
		if leftoverAmt > 0 {
			send := std.Coins{{coin.Denom, leftoverAmt}}
			banker.SendCoins(realmAddr, caller, send)
		}
	}
}

func main() {

	toSend := std.Coins{{"ugnot", 1000}}

}
