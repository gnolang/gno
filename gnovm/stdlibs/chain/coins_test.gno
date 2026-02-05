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

func TestCoins_AmountOf(t *testing.T) {
	coins := chain.Coins{
		chain.NewCoin("denom1", 1),
		chain.NewCoin("denom2", 2),
	}

	amount := coins.AmountOf("denom1")
	if amount != 1 {
		t.Fatalf("expected amount=1 for denom1, got %d", amount)
	}

	amount = coins.AmountOf("denom2")
	if amount != 2 {
		t.Fatalf("expected amount=2 for denom2, got %d", amount)
	}

	amount = coins.AmountOf("denom3")
	if amount != 0 {
		t.Fatalf("expected amount=0 for denom3, got %d", amount)
	}

	coins = append(coins, chain.NewCoin("denom1", 1))

	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Fatal("expected panic when getting duplicate denom")
			}
		}()
		coins.AmountOf("denom1")
	}()
}
