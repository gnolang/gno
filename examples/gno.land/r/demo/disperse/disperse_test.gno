package disperse

import (
	"std"
	"strings"
	"testing"
)

func TestSendCoin(t *testing.T) {
	sender := std.Address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	receiverArr := []string{"g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa", "g1x8uj9wc3wevvn3glvfallt5e5ercfcs520wy4j"}
	amountArr := []string{"100", "150"}

	// Mint test GNOT
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	// Set caller to sender
	std.TestSetOrigCaller(sender)

	std.TestSetOrigSend(std.Coins{{"ugnot", 250}}, nil)

	// Get readonly banker to check balances
	//banker := std.GetBanker(std.BankerTypeReadonly)

	addresses := strings.Join(receiverArr, ",")
	amounts := strings.Join(amountArr, ",")

	//balSenderBefore := banker.GetCoins(sender)
	SendGNOT(addresses, amounts)

	//balSenderAfter := banker.GetCoins(sender)
	//
	//println("balSenderBefore:")
	//println(balSenderBefore)
	//println("balSenderAfter:")
	//println(balSenderAfter)

}
