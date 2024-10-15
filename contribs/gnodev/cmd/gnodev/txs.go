package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func parseTxs(ctx context.Context, logger *slog.Logger, txFile string) ([]gnoland.GenesisTx, error) {
	if txFile == "" {
		return nil, nil
	}

	file, loadErr := os.Open(txFile)
	if loadErr != nil {
		return nil, fmt.Errorf("unable to open tx file %s: %w", txFile, loadErr)
	}
	defer file.Close()

	var txs []gnoland.GenesisTx
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, std.ErrTxsLoadingAborted
		default:
		}

		// XXX this can be expensive if all th txs are std.Tx
		// find a better way
		line := scanner.Bytes()

		var metatx gnoland.MetadataTx
		unmarshalMetaTxErr := amino.Unmarshal(line, &metatx)
		if unmarshalMetaTxErr == nil {
			logger.Debug("load metatx", "tx", metatx.GenesisTx, "meta", metatx.TxMetadata)
			txs = append(txs, &metatx)
			continue
		}

		// fallback on std tx
		var tx std.Tx
		unmarshalTxErr := amino.Unmarshal(line, &tx)
		if unmarshalTxErr == nil {
			logger.Debug("load tx", "tx", metatx.GenesisTx, "meta", metatx.TxMetadata)
			txs = append(txs, dev.GnoGenesisTx{StdTx: tx})
			continue
		}

		return nil, fmt.Errorf("unable to unmarshal tx: %w", errors.Join(unmarshalMetaTxErr, unmarshalTxErr))
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"error encountered while reading file, %w",
			err,
		)
	}

	return txs, nil
}
