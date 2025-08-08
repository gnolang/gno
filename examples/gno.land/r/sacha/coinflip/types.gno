package coinflip

import "std"

type Bet struct {
    Address std.Address
    Result bool
    Amount std.Coin
}
