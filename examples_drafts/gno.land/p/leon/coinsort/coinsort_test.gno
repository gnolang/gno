package coinsort

import (
	"std"
	"testing"
)

func TestSortByBalance(t *testing.T) {
	coins := std.Coins{
		std.Coin{Denom: "b", Amount: 50},
		std.Coin{Denom: "c", Amount: 10},
		std.Coin{Denom: "a", Amount: 100},
	}

	expected := std.Coins{
		std.Coin{Denom: "c", Amount: 10},
		std.Coin{Denom: "b", Amount: 50},
		std.Coin{Denom: "a", Amount: 100},
	}

	SortByBalance(coins)

	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortByBalance failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}

func TestSortByCustomDescendingAmount(t *testing.T) {
	coins := std.Coins{
		std.Coin{Denom: "a", Amount: 2},
		std.Coin{Denom: "b", Amount: 3},
		std.Coin{Denom: "c", Amount: 1},
	}

	expected := std.Coins{
		std.Coin{Denom: "b", Amount: 3},
		std.Coin{Denom: "a", Amount: 2},
		std.Coin{Denom: "c", Amount: 1},
	}

	SortBy(coins, func(a, b std.Coin) bool {
		return a.Amount > b.Amount // descending
	})

	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortBy custom descending failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}

func TestSortByNilFunc(t *testing.T) {
	coins := std.Coins{
		std.Coin{Denom: "x", Amount: 5},
		std.Coin{Denom: "z", Amount: 20},
		std.Coin{Denom: "y", Amount: 10},
	}

	expected := std.Coins{
		std.Coin{Denom: "x", Amount: 5},
		std.Coin{Denom: "z", Amount: 20},
		std.Coin{Denom: "y", Amount: 10},
	}

	SortBy(coins, nil)

	// should stay the same
	for i := range coins {
		if coins[i] != expected[i] {
			t.Errorf("SortBy nil func failed at index %d: got %+v, want %+v", i, coins[i], expected[i])
		}
	}
}
