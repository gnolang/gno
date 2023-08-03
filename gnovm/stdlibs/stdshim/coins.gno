package std

import "strconv"

// NOTE: this is selectly copied over from pkgs/std/coin.go
// TODO: import all functionality(?).

// Coin hold some amount of one currency.
// A negative amount is invalid.
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

func (c Coin) String() string {
	return strconv.Itoa(int(c.Amount)) + c.Denom
}

func (c Coin) IsGTE(other Coin) bool {
	if c.Denom != other.Denom {
		panic("invalid coin denominations: " + c.Denom)
	}
	return c.Amount >= other.Amount
}

// Coins is a set of Coin, one per currency
type Coins []Coin

func (cz Coins) String() string {
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

func (a Coins) Add(b Coins) Coins {
	c := Coins{}
	for _, ac := range a {
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

// TODO implement Coin/Coins constructors.
