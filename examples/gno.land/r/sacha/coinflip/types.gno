package coinflip

import (
	"chain"
	"chain/banker"
)

type Bet struct {
	Address chain.Address
	Result  bool
	Amount  banker.Coin
}
