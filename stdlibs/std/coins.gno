package std

import "strconv"
import "fmt"
// import "github.com/gnolang/gno/pkgs/std"

// NOTE: this is selectly copied over from pkgs/std/coin.go
// TODO: import all functionality(?).

// Coin hold some amount of one currency.
// A negative amount is invalid.
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

func (c Coin) String() string {
	return strconv.Itoa(int(c.Amount)) + "*" + c.Denom
}

func (c Coin) IsGTE(other Coin) bool {
	if c.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", c.Denom, other.Denom))
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

// TODO implement Coin/Coins constructors.
