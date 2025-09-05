package coinflip

import (
	"chain/banker"
)

type Bet struct {
	Address address
	Result  bool
	Amount  banker.Coin
}
