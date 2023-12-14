package disperse

import (
	"gno.land/p/demo/testutils"
	"std"
	"strings"
	"testing"
)

func TestSendCoin(t *testing.T) {
	// Get readonly banker to check balances
	bankerTestIssueCoin := std.GetBanker(std.BankerTypeRealmIssue)
	//cleanUpRealm(banker)

	// Get Disperse realm address
	realmAddr := std.GetOrigPkgAddr()

	// Define test addresses
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	charlie := testutils.TestAddress("bob")

	totalSendAmount := std.Coins{{"ugnot", 250}}
	//bankerRealmSend := std.GetBanker(std.BankerTypeRealmSend)

	// Mint test GNOT
	std.TestIssueCoins(alice, totalSendAmount)

	// Set tx caller to alice
	std.TestSetOrigCaller(alice)

	// Send 250ugnot in total to Disperse realm
	std.TestSetOrigSend(totalSendAmount, nil) // todo add issue for this

	//Format array for SendCoin function
	receiverArr := []string{bob.String(), charlie.String()}

	// Get initial realm balance
	//initialBalance := bankerRealmSend.GetCoins(realmAddr)

	// Get random burn address
	//zeroAddr := testutils.TestAddress("0x0")

	bankerRealmSend.SendCoins(realmAddr, "g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de", initialBalance)

	// How many tokens to send to bob & charlie
	amountArr := []string{"100", "150"} // TODO parse from & to coins properly

	addresses := strings.Join(receiverArr, ",")
	amounts := strings.Join(amountArr, ",")

	// Balance of realm should be zero
	//if initialRealmBalance := banker.GetCoins(realmAddr).AmountOf("ugnot"); initialRealmBalance != 0 {
	//	t.Fatal(errBalanceNotZero)
	//}

	// Get all balances before sending
	//aliceBalBefore := banker.GetCoins(alice).AmountOf("ugnot")
	//bobBalBefore := banker.GetCoins(bob).AmountOf("ugnot")
	//charlieBalBefore := banker.GetCoins(charlie).AmountOf("ugnot")

	// Send coin
	SendCoin(addresses, amounts)

	// Get all balances after sending
	//aliceBalAfter := banker.GetCoins(alice).AmountOf("ugnot")
	//bobBalAfter := banker.GetCoins(bob).AmountOf("ugnot")
	//charlieBalAfter := banker.GetCoins(charlie).AmountOf("ugnot")

	//if aliceBalBefore-aliceBalAfter != totalSendAmount.AmountOf("ugnot") {
	//	t.Fatal("alice's balance not correct")
	//}
}

//// Remove any existing coin balances for testing purposes
//func cleanUpRealm(banker std.BankerTypeRealmSend) {
//
//
//}
