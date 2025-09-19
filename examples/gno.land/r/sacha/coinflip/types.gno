package coinflip

import (
	"chain"
	"chain/banker"
)

type Bet struct {
	Address address
	Result  bool
	Amount  chain.Coin
}
