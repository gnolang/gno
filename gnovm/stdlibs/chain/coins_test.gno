package chain_test

import (
	"chain"
	"testing"
)

func TestCoinIsZero(t *testing.T) {
	coin := chain.NewCoin("ugnot", 0)
	res := coin.IsZero()
	if !res {
		t.Fatalf("expected IsZero=true for %s", coin)
	}

	coin = chain.NewCoin("ugnot", 1)
	res = coin.IsZero()
	if res {
		t.Fatalf("expected IsZero=false for %s", coin)
	}
}

func TestCoinsIsZero(t *testing.T) {
	var (
		testDenom1 = "denom1"
		testDenom2 = "denom2"
	)
	tests := []struct {
		name     string
		coins    chain.Coins
		expected bool
	}{
		{
			name:     "empty coins",
			coins:    chain.Coins{},
			expected: true,
		},
		{
			name:     "one zero denom",
			coins:    chain.Coins{chain.NewCoin(testDenom1, 0)},
			expected: true,
		},
		{
			name: "two zero denom",
			coins: chain.Coins{
				chain.NewCoin(testDenom1, 0), chain.NewCoin(testDenom2, 0),
			},
			expected: true,
		},
		{
			name:     "one non-zero denom",
			coins:    chain.Coins{chain.NewCoin(testDenom1, 1)},
			expected: false,
		},
		{
			name: "one zero denom, one non-zero denom",
			coins: chain.Coins{
				chain.NewCoin(testDenom1, 0), chain.NewCoin(testDenom2, 1),
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zero := tt.coins.IsZero()

			if zero != tt.expected {
				t.Fatalf("expected IsZero=%t for coins=%v", tt.expected, tt.coins)
			}
		})
	}
}

const maxInt64 int64 = (1 << 63) - 1

func shouldPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic but did not panic")
		}
	}()
	f()
}

func TestNewCoins_SortsJoinsAndRemovesZero(t *testing.T) {
	// b cancels out to zero, a sums to 4
	res := chain.NewCoins(
		chain.NewCoin("b", 2),
		chain.NewCoin("a", 1),
		chain.NewCoin("b", -2),
		chain.NewCoin("a", 3),
	)

	if len(res) != 1 {
		t.Fatalf("expected 1 coin after join, got %d: %v", len(res), res)
	}
	if res[0].Denom != "a" || res[0].Amount != 4 {
		t.Fatalf("unexpected result: %v", res)
	}
}

func TestNewCoins_ModifiesInputSliceOrdering(t *testing.T) {
	// When passing a slice with the variadic expansion (s...), the underlying
	// array elements are sorted in-place (sort.Sort operates on the shared array).
	s := []chain.Coin{
		chain.NewCoin("b", 2),
		chain.NewCoin("a", 1),
	}
	_ = chain.NewCoins(s...)

	// After NewCoins, the original slice should be sorted by denom.
	if s[0].Denom != "a" || s[1].Denom != "b" {
		t.Fatalf("expected original slice to be sorted, got: %v", s)
	}
}

func TestNewCoins_OrderBasedPanic(t *testing.T) {
	cz := chain.Coins{
		chain.NewCoin("z", 1),
		chain.NewCoin("z", -1),
		chain.NewCoin("z", maxInt64),
	}
	// Should not panic: 1 + (-1) + maxInt should not cause a panic.
	chain.NewCoins(cz...)

	cz = chain.Coins{
		chain.NewCoin("z", 1),
		chain.NewCoin("z", maxInt64),
		chain.NewCoin("z", -1),
	}
	// Should panic: 1 + maxInt causes an overflow, regardless of whether
	// it is followed up by a -1.
	shouldPanic(t, func() { chain.NewCoins(cz...) })
}

func TestCoinsAdd_Basic(t *testing.T) {
	cz := chain.NewCoins(chain.NewCoin("b", 2), chain.NewCoin("a", 1))
	other := chain.NewCoins(chain.NewCoin("a", 3), chain.NewCoin("c", 5))

	res := cz.Add(other)

	// Expect sorted result and proper joining: a:4, b:2, c:5
	if len(res) != 3 {
		t.Fatalf("expected 3 coins in result, got %d: %v", len(res), res)
	}
	if res[0].Denom != "a" || res[0].Amount != 4 {
		t.Fatalf("unexpected a coin: %v", res[0])
	}
	if res[1].Denom != "b" || res[1].Amount != 2 {
		t.Fatalf("unexpected b coin: %v", res[1])
	}
	if res[2].Denom != "c" || res[2].Amount != 5 {
		t.Fatalf("unexpected c coin: %v", res[2])
	}

	// Ensure inputs were not mutated (Add should not modify inputs)
	if cz.AmountOf("a") != 1 || cz.AmountOf("b") != 2 {
		t.Fatalf("cz was modified: %v", cz)
	}
	if other.AmountOf("a") != 3 || other.AmountOf("c") != 5 {
		t.Fatalf("other was modified: %v", other)
	}
}

func TestCoinsAdd_RemovesZero(t *testing.T) {
	// Test that sums that result in zero are removed.
	cz := chain.NewCoins(chain.NewCoin("a", 5))
	other := chain.NewCoins(chain.NewCoin("a", -5), chain.NewCoin("b", 1))

	res := cz.Add(other)
	// 'a' should be removed because 5 + (-5) == 0
	if len(res) != 1 {
		t.Fatalf("expected only 1 coin after cancellation, got %d: %v", len(res), res)
	}
	if res[0].Denom != "b" || res[0].Amount != 1 {
		t.Fatalf("unexpected remaining coin: %v", res[0])
	}
}

func TestCoinsAdd_OverflowPanicsWhenSumming(t *testing.T) {
	cz := chain.Coins{chain.NewCoin("x", maxInt64)}
	other := chain.Coins{chain.NewCoin("x", 1)}
	shouldPanic(t, func() { _ = cz.Add(other) })
}
