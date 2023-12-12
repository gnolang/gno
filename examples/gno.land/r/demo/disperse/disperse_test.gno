package disperse

import (
	"gno.land/p/demo/testutils"
	"std"
	"strings"
	"testing"
)

func TestSendCoin(t *testing.T) {
	//cleanUpRealm()

	// Define test addresses
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	charlie := testutils.TestAddress("bob")

	totalSendAmount := std.Coins{{"ugnot", 250}}
	// Mint test GNOT
	std.TestIssueCoins(alice, totalSendAmount)

	// Set tx caller to alice
	std.TestSetOrigCaller(alice)

	// Send 250ugnot in total to Disperse realm
	std.TestSetOrigSend(totalSendAmount, nil)

	// Format array for SendCoin function
	receiverArr := []string{bob.String(), charlie.String()}

	// How many tokens to send to bob & charlie
	amountArr := []string{"100", "150"} // TODO parse from & to coins properly

	addresses := strings.Join(receiverArr, ",")
	amounts := strings.Join(amountArr, ",")

	banker := std.GetBanker(std.BankerTypeReadonly)

	// Get Disperse realm address
	realmAddr := std.GetOrigPkgAddr()

	// Balance of realm should be zero
	if initialRealmBalance := banker.GetCoins(realmAddr).AmountOf("ugnot"); initialRealmBalance != 0 {
		t.Fatal(errBalanceNotZero)
	}

	// Get all balances before sending
	aliceBalBefore := banker.GetCoins(alice).AmountOf("ugnot")
	bobBalBefore := banker.GetCoins(bob).AmountOf("ugnot")
	charlieBalBefore := banker.GetCoins(charlie).AmountOf("ugnot")

	// Send coin
	SendCoin(addresses, amounts)

	// Get all balances after sending
	aliceBalAfter := banker.GetCoins(alice).AmountOf("ugnot")
	//bobBalAfter := banker.GetCoins(bob).AmountOf("ugnot")
	//charlieBalAfter := banker.GetCoins(charlie).AmountOf("ugnot")

	if aliceBalBefore-aliceBalAfter != totalSendAmount.AmountOf("ugnot") {
		t.Fatal("alice's balance not correct")
	}

	if aliceBalBefore-aliceBalAfter != totalSendAmount.AmountOf("ugnot") {
		t.Fatal("alice's balance not correct")
	}

}

// Remove any existing coin balances for testing purposes
func cleanUpRealm() {
	// Get readonly banker to check balances
	banker := std.GetBanker(std.BankerTypeRealmSend)

	// Get Disperse realm address
	realmAddr := std.GetOrigPkgAddr()

	// Get initial realm balance
	initialBalance := banker.GetCoins(realmAddr)

	// Get random burn address
	zeroAddr := testutils.TestAddress("0x0")

	banker.SendCoins(realmAddr, zeroAddr, initialBalance)
}
