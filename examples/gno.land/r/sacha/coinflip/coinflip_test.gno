package coinflip

import (
	"std"
	"testing"
)

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
	coinflipAddr := std.DerivePkgAddr("gno.land/r/sacha/coinflip")
	banker := std.NewBanker(std.BankerTypeReadonly)

	testing.IssueCoins(coinflipAddr, std.Coins{{"ugnot", 100000000}})

	sachaAddr := std.Address("g1jpg2hmsh8j4f60u4vkseqh5azeupns8vhrstfm")
	testing.SetRealm(std.NewUserRealm(sachaAddr))

	Withdrawal(cross, 900)

	mainbal := banker.GetCoins(coinflipAddr)
	sachabal := banker.GetCoins(sachaAddr)

	if !(mainbal[0].Amount == 99999100 && sachabal[0].Amount == 900) {
		t.Fatalf("Expected main balance to be 99999100 and sacha balance to be 900, got %v and %v", mainbal, sachabal)
	}
}
