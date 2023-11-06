package main

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// txStore is a wrapper for TM2 transactions
type txStore []std.Tx

// leftMerge merges the two tx stores, with
// preference to the left
func (i *txStore) leftMerge(b txStore) error {
	// Build out the tx hash map
	txHashMap := make(map[string]struct{}, len(*i))

	for _, tx := range *i {
		txHash, err := getTxHash(tx)
		if err != nil {
			return err
		}

		txHashMap[txHash] = struct{}{}
	}

	for _, tx := range b {
		txHash, err := getTxHash(tx)
		if err != nil {
			return err
		}

		if _, exists := txHashMap[txHash]; !exists {
			*i = append(*i, tx)
		}
	}

	return nil
}

type (
	accountBalances map[types.Address]int64 // address -> balance (ugnot)
	accountBalance  struct {
		address types.Address
		amount  int64
	}
)

// toList linearizes the account balances map
func (a accountBalances) toList() []string {
	balances := make([]string, 0, len(a))

	for address, balance := range a {
		balances = append(
			balances,
			fmt.Sprintf("%s=%dugnot", address, balance),
		)
	}

	return balances
}

// leftMerge left-merges the two maps
func (a accountBalances) leftMerge(b accountBalances) {
	for key, bVal := range b {
		if _, present := (a)[key]; !present {
			(a)[key] = bVal
		}
	}
}
