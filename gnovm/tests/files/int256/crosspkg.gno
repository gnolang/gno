package main

import (
	"math"
	"math/overflow"
	"strconv"
)

// int256 and uint256 are the same types in every package: values cross
// package boundaries without conversion.
type Token struct {
	Supply  uint256
	Balance map[string]uint256
}

func (t *Token) Mint(to string, amount uint256) {
	t.Supply = overflow.Addu256p(t.Supply, amount)
	t.Balance[to] = overflow.Addu256p(t.Balance[to], amount)
}

func main() {
	t := &Token{Balance: map[string]uint256{}}
	t.Mint("alice", 1e30)
	t.Mint("bob", math.MaxUint256-1e30)
	println(strconv.FormatUint256(t.Supply, 10), t.Supply == math.MaxUint256)
	defer func() {
		println("recovered:", recover())
	}()
	t.Mint("carol", 1)
}

// Output:
// 115792089237316195423570985008687907853269984665640564039457584007913129639935 true
// recovered: addition overflow
