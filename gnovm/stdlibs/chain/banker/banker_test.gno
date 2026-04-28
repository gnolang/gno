package banker

import (
	"chain"
	"testing"
)

// fakeBanker is a hand-rolled implementation of the Banker interface
// used to verify that IsCanonical rejects non-canonical bankers. It
// has no relationship to the bank-keeper.
type fakeBanker struct{}

func (fakeBanker) GetCoins(addr address) chain.Coins                   { return nil }
func (fakeBanker) SendCoins(from, to address, amt chain.Coins)         {}
func (fakeBanker) TotalCoin(denom string) int64                        { return 0 }
func (fakeBanker) IssueCoin(addr address, denom string, amount int64)  {}
func (fakeBanker) RemoveCoin(addr address, denom string, amount int64) {}

func TestIsCanonical_CanonicalValue(t *testing.T) {
	// Canonical type constructed via in-package struct literal. Outside
	// the package, this construction is impossible (banker is unexported)
	// — only NewBanker can produce one — so this same-package test stands
	// in for "the value NewBanker would have returned."
	b := banker{bt: BankerTypeRealmSend, pkgAddr: address("g1foo")}
	if !IsCanonical(b) {
		t.Fatal("expected canonical banker to satisfy IsCanonical")
	}
}

func TestIsCanonical_Nil(t *testing.T) {
	if IsCanonical(nil) {
		t.Fatal("expected nil Banker to NOT satisfy IsCanonical")
	}
}

func TestIsCanonical_HandRolledFake(t *testing.T) {
	if IsCanonical(fakeBanker{}) {
		t.Fatal("expected hand-rolled fakeBanker to NOT satisfy IsCanonical")
	}
}
