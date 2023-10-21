package main

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

type (
	accountBalances map[types.Address]int64 // address -> balance (ugnot)
	accountBalance  struct {
		address types.Address
		amount  int64
	}
)

// toList linearizes the account balances map
func (a *accountBalances) toList() []string {
	balances := make([]string, 0, len(*a))

	for address, balance := range *a {
		balances = append(
			balances,
			fmt.Sprintf("%s=%dugnot", address, balance),
		)
	}

	return balances
}

// leftMerge left-merges the two maps
func (a *accountBalances) leftMerge(b accountBalances) {
	for key, bVal := range b {
		if _, present := (*a)[key]; !present {
			(*a)[key] = bVal
		}
	}
}
