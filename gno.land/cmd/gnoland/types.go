package main

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
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

type accountBalances map[types.Address]gnoland.Balance // address -> balance (ugnot)

// toList linearizes the account balances map
func (a accountBalances) toList() []gnoland.Balance {
	balances := make([]gnoland.Balance, 0, len(a))

	for _, balance := range a {
		balances = append(balances, balance)
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
