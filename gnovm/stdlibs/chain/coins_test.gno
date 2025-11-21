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
				t.Fatal("exepected IsZero=%t for coins=%s", tt.expected, tt.coins)
			}
		})
	}
}
