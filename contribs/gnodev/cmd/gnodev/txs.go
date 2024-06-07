package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func parseTxs(txFile string) ([]std.Tx, error) {
	if txFile == "" {
		return nil, nil
	}

	file, loadErr := os.Open(txFile)
	if loadErr != nil {
		return nil, fmt.Errorf("unable to open tx file %s: %w", txFile, loadErr)
	}
	defer file.Close()

	return std.ParseTxs(context.Background(), file)
}
