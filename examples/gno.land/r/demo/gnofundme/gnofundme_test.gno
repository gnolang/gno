package gnofundme

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/r/demo/gnofundme"
)

func TestWithdraw(t *testing.T) {
	// Setup
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/gnofundme")
	mainaddr := std.DerivePkgAddr("main")
	std.TestSetOrigCaller(mainaddr)
	std.TestSetOrigPkgAddr(banktestAddr)

	// get and print balance of mainaddr.
	// with the SEND, + 200 gnot given by the TestContext, main should have 300gnot.
	banker := std.GetBanker(std.BankerTypeRealmSend)
	mainbal := banker.GetCoins(mainaddr)
	ctype := 0
	title := "Campaign for Withdrawal"
	description := "Description for Withdrawal"
	goal := int64(1000)
	begin := time.Now()
	deadline := begin.Add(time.Second * 1)
	ownerAddress := testutils.TestAddress("owner")
	beneficiaryAddress := testutils.TestAddress("beneficiary")
	std.TestSetOrigCaller(beneficiaryAddress)
	NewUser("Beneficiary", "")
	std.TestSetOrigCaller(ownerAddress)
	NewUser("Owner", "")
	NewCampaign(ctype, title, description, goal, begin.Unix(), deadline.Unix(), beneficiaryAddress)

	// Add contributions to the campaign
	contributorAddress := testutils.TestAddress("contributor")
	std.TestSetOrigCaller(contributorAddress)
	NewUser("Contributor", "")
	var amount int64 = 500
	display := true
	// date := time.Unix(time.Now().Unix(), 0)
	banker.SendCoins(banktestAddr, mainaddr, std.Coins{{"ugnot", 500}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 500}}, nil)
	AddContributor("0", display)

	// Fast forward time to after the campaign deadline
	// Simulate time passing
	// time.Sleep(time.Second * 2)
	// deadline = time.Now().Add(-time.Second * 1)
	std.TestSetOrigCaller(ownerAddress)
	std.TestSkipHeights(1)
	// Withdraw funds
	Withdraw("0")

	// Verify
	if campaignsTotal["0"] != 0 {
		t.Errorf("Expected campaignsTotal to be 0 after withdrawal, got %d", campaignsTotal["0"])
	}

	beneficiaryBalance := banker.GetCoins(beneficiaryAddress).AmountOf("ugnot")
	if beneficiaryBalance != amount {
		t.Errorf("Expected beneficiary balance to be %d, got %d", amount, beneficiaryBalance)
	}
}
