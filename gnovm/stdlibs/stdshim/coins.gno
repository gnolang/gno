package std

import (
	"fmt"
	"regexp"
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
// It will panic if the amount is negative.
// To construct a negative (invalid) amount, use an operation.
func NewCoin(denom string, amount int64) Coin {
	if err := validate(denom, amount); err != nil {
		panic(err)
	}

	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}

// String returns a stringified version of Coin, i.e. "100ugnot"
func (c Coin) String() string {
	return strconv.Itoa(int(c.Amount)) + c.Denom
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (c Coin) IsGTE(other Coin) bool {
	if c.Denom != other.Denom {
		panic("invalid coin denominations; " + c.Denom + ", " + other.Denom)
	}
	return c.Amount >= other.Amount
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value
func (c Coin) IsLT(other Coin) bool {
	if c.Denom != other.Denom {
		panic("invalid coin denominations; " + c.Denom + ", " + other.Denom)
	}
	return c.Amount < other.Amount
}

// IsEqual returns true if the two sets of Coins have the same value
func (c Coin) IsEqual(other Coin) bool {
	if c.Denom != other.Denom {
		panic("invalid coin denominations; " + c.Denom + ", " + other.Denom)
	}
	return c.Amount == other.Amount
}

// validate returns an error if the Coin has a negative amount or if
// the denom is invalid.
func validate(denom string, amount int64) error {
	if err := validateDenom(denom); err != nil {
		return err
	}

	if amount < 0 {
		return fmt.Errorf("negative coin amount: %d", amount)
	}

	return nil
}

// IsValid returns true if the Coin has a non-negative amount and the denom is valid.
func (c Coin) IsValid() bool {
	if err := validate(c.Denom, c.Amount); err != nil {
		return false
	}
	return true
}

func (c Coin) IsZero() bool {
	return c.Amount == 0
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

func validateDenom(denom string) error {
	if !reDnm.MatchString(denom) {
		return fmt.Errorf("invalid denom: %s", denom)
	}
	return nil
}

var (
	// Denominations can be 3 ~ 16 characters long.
	reDnmString = `[a-z][a-z0-9]{2,15}`
	reDnm       = regexp.MustCompile(fmt.Sprintf(`^%s$`, reDnmString))
)
