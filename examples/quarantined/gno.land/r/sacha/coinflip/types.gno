package coinflip

import (
	"chain"
)

type Bet struct {
	Address address
	Result  bool
	Amount  chain.Coin
}
