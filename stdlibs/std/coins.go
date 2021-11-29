package std

// NOTE: this is selectly copied over from pkgs/std/coin.go
// TODO: import all functionality(?).

// Coin hold some amount of one currency.
// A negative amount is invalid.
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

// Coins is a set of Coin, one per currency
type Coins []Coin
