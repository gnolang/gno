package client

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/gas"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/ryanuber/columnize"
	"github.com/xlab/treeprint"
)

// PrintTxInfo prints the transaction result to io. If the events has storage deposit
// info then also print it with the total transaction cost.
func PrintTxInfo(tx std.Tx, res *ctypes.ResultBroadcastTxCommit, io commands.IO, verbosity int) {
	io.Println(string(res.DeliverTx.Data))
	io.Println("OK!")
	io.Println("GAS WANTED:", res.DeliverTx.GasWanted)

	if verbosity == 0 {
		io.Println("GAS USED:  ", res.DeliverTx.GasUsed.Total.GasConsumed)
	} else {
		printGasDetail(res.DeliverTx.GasUsed, io, verbosity)
	}

	io.Println("HEIGHT:    ", res.Height)
	if bytesDelta, coinsDelta, hasStorageEvents := GetStorageInfo(res.DeliverTx.Events); hasStorageEvents {
		io.Printfln("STORAGE DELTA:  %d bytes", bytesDelta)
		if coinsDelta.IsAllPositive() || coinsDelta.IsZero() {
			io.Println("STORAGE FEE:   ", coinsDelta)
		} else {
			// NOTE: there is edge cases where coinsDelta can be a mixture of positive and negative coins.
			// For example if the keeper respects the storage price param denom and a tx contains a storage cost param change message sandwiched by storage movement messages.
			// These will fall in this case and print confusing information but it's so rare that we don't
			// really care about this possibility here.
			io.Println("STORAGE REFUND:", std.Coins{}.SubUnsafe(coinsDelta))
		}
		io.Printfln("TOTAL TX COST:  %s", coinsDelta.AddUnsafe(std.Coins{tx.Fee.GasFee}))
	}
	io.Println("EVENTS:    ", string(res.DeliverTx.EncodeEvents()))
	io.Println("INFO:      ", res.DeliverTx.Info)
	io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(res.Hash))
}

// printGasDetail prints detailed gas usage information based on the verbosity level.
func printGasDetail(detail gas.GasDetail, io commands.IO, verbosity int) {
	// Helper function to format each line with proper columns.
	delimitColumns := func(name string, value gas.Detail) string {
		return fmt.Sprintf(
			"%s: | Operation: %d | Gas: %d",
			name,
			value.OperationCount,
			value.GasConsumed,
		)
	}

	tree := treeprint.NewWithRoot(delimitColumns("GAS USED", detail.Total))
	categoryDetails := detail.CategoryDetails()

	// Get sorted category map keys to iterate in order.
	categoryKeys := make([]string, 0, len(categoryDetails))
	for key := range categoryDetails {
		categoryKeys = append(categoryKeys, key)
	}
	sort.Strings(categoryKeys)

	for _, categoryKey := range categoryKeys {
		category := categoryDetails[categoryKey]

		// Add operation details only if verbosity is 2 or higher.
		if verbosity >= 2 {
			// Skip categories with zero total operation count unless verbosity is 3.
			if category.Total.OperationCount == 0 && verbosity != 3 {
				continue
			}

			branch := tree.AddBranch(delimitColumns(categoryKey, category.Total))

			// Get sorted operation map keys to iterate in order.
			operationKeys := make([]int, 0, len(category.Operations))
			for key := range category.Operations {
				operationKeys = append(operationKeys, int(key))
			}
			sort.Ints(operationKeys)

			for _, operationKey := range operationKeys {
				operation := gas.Operation(operationKey)
				operationDetail := category.Operations[operation]

				// Skip operations with zero operation count unless verbosity is 3.
				if operationDetail.OperationCount == 0 && verbosity != 3 {
					continue
				}
				branch.AddNode(delimitColumns(operation.String(), operationDetail))
			}
			// Else add only category total if there were any operations.
		} else if category.Total.OperationCount > 0 {
			tree.AddBranch(delimitColumns(categoryKey, category.Total))
		}
	}

	// Render the tree as separated lines.
	lines := strings.Split(tree.String(), "\n")

	// Format the lines into aligned columns and print.
	config := columnize.DefaultConfig()
	config.NoTrim = true
	io.Println(columnize.Format(lines, config))
}

// GetStorageInfo searches events for StorageDepositEvent or StorageUnlockEvent and returns the bytes delta and coins delta. The coins delta omits RefundWithheld.
func GetStorageInfo(events []abci.Event) (int64, std.Coins, bool) {
	var (
		bytesDelta int64
		coinsDelta std.Coins
		hasEvents  bool
	)

	for _, event := range events {
		switch storageEvent := event.(type) {
		case chain.StorageDepositEvent:
			bytesDelta += storageEvent.BytesDelta
			coinsDelta = coinsDelta.AddUnsafe(std.Coins{storageEvent.FeeDelta})
			hasEvents = true
		case chain.StorageUnlockEvent:
			bytesDelta += storageEvent.BytesDelta
			if !storageEvent.RefundWithheld {
				coinsDelta = coinsDelta.SubUnsafe(std.Coins{storageEvent.FeeRefund})
			}
			hasEvents = true
		}
	}

	return bytesDelta, coinsDelta, hasEvents
}
