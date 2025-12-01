package subscriptions

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/nt/testutils"
)

func TestCoins(t *testing.T) {
	native := chain.NewCoins(
		chain.NewCoin("ugnot", 1000),
	)
	grc20 := chain.NewCoins(
		chain.NewCoin("gno.land/r/demo/foo20", 500),
	)
	expected := chain.NewCoins(
		chain.NewCoin("/native/ugnot", 1000),
		chain.NewCoin("/grc20/gno.land/r/demo/foo20", 500),
	)

	result, err := Coins(cross, native, grc20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, coin := range expected {
		if result[i].Amount != coin.Amount {
			t.Fatalf("invalid amount of %s got %d, want %d", coin.Denom, result[i].Amount, coin.Amount)
		}
		if result[i].Denom != coin.Denom {
			t.Fatalf("invalid denom: got %s, want %s", result[i].Denom, coin.Denom)
		}
	}
}

func TestCoinsHasPositive(t *testing.T) {
	tests := []struct {
		coins    chain.Coins
		expected bool
	}{
		{chain.NewCoins(chain.NewCoin("coinA", 0)), false},
		{chain.NewCoins(chain.NewCoin("coinA", -100)), false},
		{chain.NewCoins(chain.NewCoin("coinA", 50)), true},
		{chain.NewCoins(chain.NewCoin("coinA", 0), chain.NewCoin("coinB", 0)), false},
		{chain.NewCoins(chain.NewCoin("coinA", -10), chain.NewCoin("coinB", 20)), true},
	}

	for _, test := range tests {
		result := coinsHasPositive(test.coins)
		if result != test.expected {
			t.Fatalf("coinsHasPositive(%v) = %v; want %v", test.coins, result, test.expected)
		}
	}
}

func TestAddCoins(t *testing.T) {
	a := chain.NewCoins(
		chain.NewCoin("coinA", 100),
		chain.NewCoin("coinB", 200),
	)
	b := chain.NewCoins(
		chain.NewCoin("coinB", 150),
		chain.NewCoin("coinC", 300),
	)
	expected := chain.NewCoins(
		chain.NewCoin("coinA", 100),
		chain.NewCoin("coinB", 350),
		chain.NewCoin("coinC", 300),
	)

	result := addCoins(a, b)
	if len(result) != len(expected) {
		t.Fatalf("invalid result length: got %d, want %d", len(result), len(expected))
	}
	for _, coin := range expected {
		found := false
		for _, resCoin := range result {
			if resCoin.Denom == coin.Denom {
				found = true
				if resCoin.Amount != coin.Amount {
					t.Fatalf("invalid amount for %s: got %d, want %d", coin.Denom, resCoin.Amount, coin.Amount)
				}
				break
			}
		}
		if !found {
			t.Fatalf("missing coin in result: %s", coin.Denom)
		}
	}
}

func TestSubCoins(t *testing.T) {
	a := chain.NewCoins(
		chain.NewCoin("coinA", 300),
		chain.NewCoin("coinB", 400),
	)
	b := chain.NewCoins(
		chain.NewCoin("coinB", 150),
		chain.NewCoin("coinC", 100),
	)
	expected := chain.NewCoins(
		chain.NewCoin("coinA", 300),
		chain.NewCoin("coinB", 250),
		chain.NewCoin("coinC", -100),
	)

	result := subCoins(a, b)
	if len(result) != len(expected) {
		t.Fatalf("invalid result length: got %d, want %d", len(result), len(expected))
	}
	for _, coin := range expected {
		found := false
		for _, resCoin := range result {
			if resCoin.Denom == coin.Denom {
				found = true
				if resCoin.Amount != coin.Amount {
					t.Fatalf("invalid amount for %s: got %d, want %d", coin.Denom, resCoin.Amount, coin.Amount)
				}
				break
			}
		}
		if !found {
			t.Fatalf("missing coin in result: %s", coin.Denom)
		}
	}
}

func TestMultiplyCoins(t *testing.T) {
	coins := chain.NewCoins(
		chain.NewCoin("coinA", 100),
		chain.NewCoin("coinB", 200),
	)
	factor := int64(3)
	expected := chain.NewCoins(
		chain.NewCoin("coinA", 300),
		chain.NewCoin("coinB", 600),
	)

	result := multiplyCoins(coins, factor)
	if len(result) != len(expected) {
		t.Fatalf("invalid result length: got %d, want %d", len(result), len(expected))
	}
	for _, coin := range expected {
		found := false
		for _, resCoin := range result {
			if resCoin.Denom == coin.Denom {
				found = true
				if resCoin.Amount != coin.Amount {
					t.Fatalf("invalid amount for %s: got %d, want %d", coin.Denom, resCoin.Amount, coin.Amount)
				}
				break
			}
		}
		if !found {
			t.Fatalf("missing coin in result: %s", coin.Denom)
		}
	}
}

func TestAddCoinAmount(t *testing.T) {
	coins := chain.NewCoins(
		chain.NewCoin("coinA", 100),
		chain.NewCoin("coinB", 200),
	)
	value := chain.NewCoin("coinB", 150)
	expected := chain.NewCoins(
		chain.NewCoin("coinA", 100),
		chain.NewCoin("coinB", 350),
	)

	result := addCoinAmount(coins, value)
	if len(result) != len(expected) {
		t.Fatalf("invalid result length: got %d, want %d", len(result), len(expected))
	}
	for _, coin := range expected {
		found := false
		for _, resCoin := range result {
			if resCoin.Denom == coin.Denom {
				found = true
				if resCoin.Amount != coin.Amount {
					t.Fatalf("invalid amount for %s: got %d, want %d", coin.Denom, resCoin.Amount, coin.Amount)
				}
				break
			}
		}
		if !found {
			t.Fatalf("missing coin in result: %s", coin.Denom)
		}
	}
}

func TestSendCoins(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	coinsToSend := chain.NewCoins(
		chain.NewCoin("/native/gno.land/r/samcrew/subscriptions:a", 100),
		chain.NewCoin("/native/gno.land/r/samcrew/subscriptions:b", 200),
	)

	// without native prefix
	coinsToIssue := chain.NewCoins(
		chain.NewCoin("gno.land/r/samcrew/subscriptions:a", 100),
		chain.NewCoin("gno.land/r/samcrew/subscriptions:b", 200),
	)

	testing.IssueCoins(alice, coinsToIssue)

	bob := testutils.TestAddress("bob")
	testing.SetOriginSend(coinsToSend)

	sendCoins(bob, coinsToSend)

	banker := banker.NewBanker(banker.BankerTypeRealmIssue)
	aliceCoins := banker.GetCoins(alice)
	bobCoins := banker.GetCoins(bob)
	for _, coin := range coinsToIssue {
		aliceAmount := int64(0)
		bobAmount := int64(0)
		for _, aCoin := range aliceCoins {
			if aCoin.Denom == coin.Denom {
				aliceAmount = aCoin.Amount
				break
			}
		}
		for _, bCoin := range bobCoins {
			if bCoin.Denom == coin.Denom {
				bobAmount = bCoin.Amount
				break
			}
		}

		expectedAliceAmount := int64(0)
		expectedBobAmount := coin.Amount
		if aliceAmount != expectedAliceAmount {
			t.Fatalf("invalid alice amount for %s: got %d, want %d", coin.Denom, aliceAmount, expectedAliceAmount)
		}
		if bobAmount != expectedBobAmount {
			t.Fatalf("invalid bob amount for %s: got %d, want %d", coin.Denom, bobAmount, expectedBobAmount)
		}
	}
}
