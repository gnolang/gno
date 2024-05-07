package main

import (
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
