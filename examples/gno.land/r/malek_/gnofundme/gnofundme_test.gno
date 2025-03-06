package gnofundme

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	"gno.land/r/demo/users"
)

func TestWithdraw(t *testing.T) {
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/gnofundme")
	mainaddr := std.DerivePkgAddr("main")
	std.TestSetOrigCaller(mainaddr)
	std.TestSetOrigPkgAddr(banktestAddr)

	banker := std.GetBanker(std.BankerTypeRealmSend)
	owner := users.GetUserByName("demo")
	beneficiary := users.GetUserByName("test1")
	std.TestSetOrigCaller(owner.Address)
	NewCampaign(0, "Campaign for Withdrawal", "Description for Withdrawal", 1000, time.Now().Unix(), time.Now().Add(time.Second*1).Unix(), beneficiary.Address)

	// Add a contribution to the campaign by owner
	var amount int64 = 500
	display := true
	std.TestSetOrigSend(std.Coins{{"ugnot", amount}}, nil)
	AddContributor("0", display)
	std.TestSkipHeights(1)
	// Withdraw funds
	Withdraw("0")

	// Verify
	uassert.Equal(t, campaignsTotal["0"], int64(0))

	beneficiaryBalance := banker.GetCoins(beneficiary.Address).AmountOf("ugnot")
	uassert.Equal(t, beneficiaryBalance, amount)
}
