package std

import (
	"math/overflow"
	"strconv"
)

// NOTE: this is selectively copied over from tm2/pkgs/std/coin.go

// Coin hold some amount of one currency.
// A negative amount is invalid.
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

// NewCoin returns a new coin with a denomination and amount
func NewCoin(denom string, amount int64) Coin {
	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}

// String provides a human-readable representation of a coin
func (c Coin) String() string {
	return strconv.Itoa(int(c.Amount)) + c.Denom
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (c Coin) IsGTE(other Coin) bool {
	mustMatchDenominations(c.Denom, other.Denom)

	return c.Amount >= other.Amount
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value
func (c Coin) IsLT(other Coin) bool {
	mustMatchDenominations(c.Denom, other.Denom)

	return c.Amount < other.Amount
}

// IsEqual returns true if the two sets of Coins have the same value
func (c Coin) IsEqual(other Coin) bool {
	mustMatchDenominations(c.Denom, other.Denom)

	return c.Amount == other.Amount
}

// Adds amounts of two coins with same denom.
// If the coins differ in denom then it panics.
// An overflow or underflow panics.
// An invalid result panics.
func (c Coin) Add(coinB Coin) Coin {
	res := c.AddUnsafe(coinB)
	return res
}

func (c Coin) AddUnsafe(coinB Coin) Coin {
	mustMatchDenominations(c.Denom, coinB.Denom)

	sum, ok := overflow.Add64(c.Amount, coinB.Amount)
	if !ok {
		panic("coin add overflow/underflow: " + strconv.Itoa(int(c.Amount)) + " + " + strconv.Itoa(int(coinB.Amount)))
	}

	c.Amount = sum
	return c
}

// Subtracts amounts of two coins with same denom.
// If the coins differ in denom then it panics.
// An overflow or underflow panics.
// An invalid result panics.
func (c Coin) Sub(coinB Coin) Coin {
	res := c.SubUnsafe(coinB)
	return res
}

func (c Coin) SubUnsafe(coinB Coin) Coin {
	mustMatchDenominations(c.Denom, coinB.Denom)

	dff, ok := overflow.Sub64(c.Amount, coinB.Amount)
	if !ok {
		panic("coin add overflow/underflow: " + strconv.Itoa(int(c.Amount)) + " + " + strconv.Itoa(int(coinB.Amount)))
	}
	c.Amount = dff

	return c
}

// IsPositive returns true if coin amount is positive.
func (c Coin) IsPositive() bool {
	return c.Amount > 0
}

// IsNegative returns true if the coin amount is negative and false otherwise.
func (c Coin) IsNegative() bool {
	return c.Amount < 0
}

// IsZero returns if this represents no money
func (c Coin) IsZero() bool {
	return c.Amount == 0
}

func mustMatchDenominations(denomA, denomB string) {
	if denomA != denomB {
		panic("incompatible coin denominations: " + denomA + ", " + denomB)
	}
}

// Coins is a set of Coin, one per currency
type Coins []Coin

func NewCoins(coins ...Coin) Coins {
	return coins
}

func (cz Coins) String() string {
	if len(cz) == 0 {
		return ""
	}

	res := ""
	for i, c := range cz {
		if i > 0 {
			res += ","
		}
		res += c.String()
	}

	return res
}

func (cz Coins) AmountOf(denom string) int64 {
	for _, c := range cz {
		if c.Denom == denom {
			return c.Amount
		}
	}

	return 0
}

func (cz Coins) Add(b Coins) Coins {
	c := Coins{}
	for _, ac := range cz {
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

// expand for usage within natively bound functions.
func (cz Coins) expandNative() (denoms []string, amounts []int64) {
	denoms = make([]string, len(cz))
	amounts = make([]int64, len(cz))
	for i, coin := range cz {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amount
	}

	return denoms, amounts
}
