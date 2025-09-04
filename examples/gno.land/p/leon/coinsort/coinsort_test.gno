package coinsort

import (
	"chain/banker"
	"testing"
)

func TestSortByBalance(t *testing.T) {
	coins := banker.Coins{
		banker.Coin{Denom: "b", Amount: 50},
		banker.Coin{Denom: "c", Amount: 10},
		banker.Coin{Denom: "a", Amount: 100},
	}

	expected := banker.Coins{
		banker.Coin{Denom: "c", Amount: 10},
		banker.Coin{Denom: "b", Amount: 50},
		banker.Coin{Denom: "a", Amount: 100},
	}

	SortByBalance(coins)

	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortByBalance failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}

func TestSortByCustomDescendingAmount(t *testing.T) {
	coins := banker.Coins{
		banker.Coin{Denom: "a", Amount: 2},
		banker.Coin{Denom: "b", Amount: 3},
		banker.Coin{Denom: "c", Amount: 1},
	}

	expected := banker.Coins{
		banker.Coin{Denom: "b", Amount: 3},
		banker.Coin{Denom: "a", Amount: 2},
		banker.Coin{Denom: "c", Amount: 1},
	}

	SortBy(coins, func(a, b banker.Coin) bool {
		return a.Amount > b.Amount // descending
	})

	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortBy custom descending failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}

func TestSortByNilFunc(t *testing.T) {
	coins := banker.Coins{
		banker.Coin{Denom: "x", Amount: 5},
		banker.Coin{Denom: "z", Amount: 20},
		banker.Coin{Denom: "y", Amount: 10},
	}

	expected := banker.Coins{
		banker.Coin{Denom: "x", Amount: 5},
		banker.Coin{Denom: "z", Amount: 20},
		banker.Coin{Denom: "y", Amount: 10},
	}

	SortBy(coins, nil)

	// should stay the same
	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortBy nil func failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}
