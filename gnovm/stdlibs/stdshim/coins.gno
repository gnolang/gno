package std

import (
	"sort"
	"strconv"
)

// NOTE: this is selectly copied over from pkgs/std/coin.go
// TODO: import all functionality(?).

// Coin hold some amount of one currency.
// A negative amount is invalid.
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

// NewCoin returns a new coin with a denomination and amount.
func NewCoin(denom string, amount int64) Coin {
	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}

func (coin Coin) String() string {
	return strconv.Itoa(int(coin.Amount)) + coin.Denom
}

func (coin Coin) IsGTE(other Coin) bool {
	if coin.Denom != other.Denom {
		panic("invalid coin denominations: " + coin.Denom)
	}
	return coin.Amount >= other.Amount
}

// IsZero returns if this represents no money
func (coin Coin) IsZero() bool {
	return coin.Amount == 0
}

// Coins is a set of Coin, one per currency
type Coins []Coin

// NewCoins constructs a new coin set.
func NewCoins(coins ...Coin) Coins {
	// remove zeroes
	newCoins := removeZeroCoins(Coins(coins))
	if len(newCoins) == 0 {
		return Coins{}
	}

	newCoins.Sort()

	// detect duplicate Denoms
	if dupIndex := findDup(newCoins); dupIndex != -1 {
		panic("found duplicate denom when constructing coins set")
	}
	return newCoins
}

func (coins Coins) String() string {
	res := ""
	for i, c := range coins {
		if i > 0 {
			res += ","
		}
		res += c.String()
	}
	return res
}

func (coins Coins) AmountOf(denom string) int64 {
	for _, c := range coins {
		if c.Denom == denom {
			return c.Amount
		}
	}
	return 0
}

func (coins Coins) Add(b Coins) Coins {
	c := Coins{}
	for _, ac := range coins {
		bc := b.AmountOf(ac.Denom)
		ac.Amount += bc
		c = append(c, ac)
	}
	for _, bc := range b {
		cc := c.AmountOf(bc.Denom)
		if cc == 0 {
			c = append(c, bc)
		}
	}
	return c
}

// removeZeroCoins removes all zero coins from the given coin set in-place.
func removeZeroCoins(coins Coins) Coins {
	i, l := 0, len(coins)
	for i < l {
		if coins[i].IsZero() {
			// remove coin
			coins = append(coins[:i], coins[i+1:]...)
			l--
		} else {
			i++
		}
	}

	return coins[:i]
}

// findDup works on the assumption that coins is sorted
func findDup(coins Coins) int {
	if len(coins) <= 1 {
		return -1
	}

	prevDenom := coins[0].Denom
	for i := 1; i < len(coins); i++ {
		if coins[i].Denom == prevDenom {
			return i
		}
		prevDenom = coins[i].Denom
	}

	return -1
}

// -----------------------------------------------------------------------------
// Sort interface

func (coins Coins) Len() int           { return len(coins) }
func (coins Coins) Less(i, j int) bool { return coins[i].Denom < coins[j].Denom }
func (coins Coins) Swap(i, j int)      { coins[i], coins[j] = coins[j], coins[i] }

var _ sort.Interface = Coins{}

// Sort is a helper function to sort the set of coins inplace
func (coins Coins) Sort() Coins {
	sort.Sort(coins)
	return coins
}
