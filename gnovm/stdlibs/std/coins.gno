package std

import (
	"strconv"
)

// NOTE: this is selectly copied over from pkgs/std/coin.go
// TODO: import all functionality(?).
// TODO: implement Coin/Coins constructors.

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
	if c.IsZero() {
		return ""
	}

	return strconv.Itoa(int(c.Amount)) + c.Denom
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (c Coin) IsGTE(other Coin) bool {
	if c.Denom != other.Denom {
		panic("comparing different denominations: " + c.Denom + ", " + other.Denom)
	}
	return c.Amount >= other.Amount
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value
func (c Coin) IsLT(other Coin) bool {
	if c.Denom != other.Denom {
		panic("comparing different denominations: " + c.Denom + ", " + other.Denom)
	}
	return c.Amount < other.Amount
}

// IsEqual returns true if the two sets of Coins have the same value
func (c Coin) IsEqual(other Coin) bool {
	if c.Denom != other.Denom {
		panic("comparing different denominations: " + c.Denom + ", " + other.Denom)
	}
	return c.Amount == other.Amount
}

// IsZero returns if this represents no money
func (c Coin) IsZero() bool {
	return c.Amount == 0
}

// Coins is a set of Coin, one per currency
type Coins []Coin

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
