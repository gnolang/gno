package coinflip

import (
	"std"
	"testing"

	"gno.land/p/nt/testutils"
)

var coinflipAddr = std.DerivePkgAddr("gno.land/r/sacha/coinflip")

func TestGetRealmBalance(t *testing.T) {
	value := GetRealmBalance(cross)

	if value != 0 {
		t.Fatalf("Expected 0, got %v", value)
	}
}

func TestFlip(t *testing.T) {
	value := flip(1)

	if value != true {
		t.Fatalf("Expected true, got %v", value)
	}

	value = flip(0)

	if value != false {
		t.Fatalf("Expected false, got %v", value)
	}
}

func TestWithdrawal(t *testing.T) {
	banker := std.NewBanker(std.BankerTypeReadonly)
	sachaAddr := std.Address("g1jpg2hmsh8j4f60u4vkseqh5azeupns8vhrstfm")

	testing.IssueCoins(coinflipAddr, std.Coins{{"ugnot", 100000000}})

	testing.SetRealm(std.NewUserRealm(sachaAddr))

	Withdrawal(cross, 1000)

	mainbal := banker.GetCoins(coinflipAddr)
	sachabal := banker.GetCoins(sachaAddr)

	if !(mainbal[0].Amount == 99999000 && sachabal[0].Amount == 1000) {
		t.Fatalf("Expected main balance to be 99999100 and sacha balance to be 900, got %v and %v", mainbal, sachabal)
	}
}

func TestSendPrize(t *testing.T) {
	banker := std.NewBanker(std.BankerTypeReadonly)
	aliceAddr := testutils.TestAddress("alice")

	testing.IssueCoins(coinflipAddr, std.Coins{{"ugnot", 100000000}})

	testing.SetRealm(std.NewUserRealm(aliceAddr))

	sendPrize(cross, aliceAddr, 1000)

	mainbal := banker.GetCoins(coinflipAddr)
	alicebal := banker.GetCoins(aliceAddr)
	if !(mainbal[0].Amount == 99999000 && alicebal[0].Amount == 1000) {
		t.Fatalf("Expected main balance to be 99999100 and alice balance to be 900, got %v and %v", mainbal, alicebal)
	}
}
