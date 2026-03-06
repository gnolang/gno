package coinsort

import (
	"chain"
	"testing"
)

func TestSortByBalance(t *testing.T) {
	coins := chain.Coins{
		chain.Coin{Denom: "b", Amount: 50},
		chain.Coin{Denom: "c", Amount: 10},
		chain.Coin{Denom: "a", Amount: 100},
	}

	expected := chain.Coins{
		chain.Coin{Denom: "c", Amount: 10},
		chain.Coin{Denom: "b", Amount: 50},
		chain.Coin{Denom: "a", Amount: 100},
	}

	SortByBalance(coins)

	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortByBalance failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}

func TestSortByCustomDescendingAmount(t *testing.T) {
	coins := chain.Coins{
		chain.Coin{Denom: "a", Amount: 2},
		chain.Coin{Denom: "b", Amount: 3},
		chain.Coin{Denom: "c", Amount: 1},
	}

	expected := chain.Coins{
		chain.Coin{Denom: "b", Amount: 3},
		chain.Coin{Denom: "a", Amount: 2},
		chain.Coin{Denom: "c", Amount: 1},
	}

	SortBy(coins, func(a, b chain.Coin) bool {
		return a.Amount > b.Amount // descending
	})

	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortBy custom descending failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}

func TestSortByNilFunc(t *testing.T) {
	coins := chain.Coins{
		chain.Coin{Denom: "x", Amount: 5},
		chain.Coin{Denom: "z", Amount: 20},
		chain.Coin{Denom: "y", Amount: 10},
	}

	expected := chain.Coins{
		chain.Coin{Denom: "x", Amount: 5},
		chain.Coin{Denom: "z", Amount: 20},
		chain.Coin{Denom: "y", Amount: 10},
	}

	SortBy(coins, nil)

	// should stay the same
	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortBy nil func failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}
