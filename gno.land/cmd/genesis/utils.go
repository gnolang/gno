package main

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func getTxHash(tx std.Tx) (string, error) {
	encodedTx, err := amino.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("unable to marshal transaction, %w", err)
	}

	txHash := types.Tx(encodedTx).Hash()

	return fmt.Sprintf("%X", txHash), nil
}
