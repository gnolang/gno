package main

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type txStore []std.Tx

func (i *txStore) leftMerge(b txStore) error {
	// Build out the tx hash map
	txHashMap := make(map[string]struct{}, len(*i))

	getTxHash := func(tx std.Tx) (string, error) {
		encodedTx, err := amino.Marshal(tx)
		if err != nil {
			return "", fmt.Errorf("unable to marshal transaction, %w", err)
		}

		txHash := types.Tx(encodedTx).Hash()

		return fmt.Sprintf("%X", txHash), nil
	}

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
