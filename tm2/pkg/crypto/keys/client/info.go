package client

import (
	"encoding/base64"

	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// PrintTxInfo prints the transaction result to io. If the events has storage deposit
// info then also print it with the total transaction cost.
func PrintTxInfo(tx std.Tx, res *ctypes.ResultBroadcastTxCommit, io commands.IO) {
	io.Println(string(res.DeliverTx.Data))
	io.Println("OK!")
	io.Println("GAS WANTED:", res.DeliverTx.GasWanted)
	io.Println("GAS USED:  ", res.DeliverTx.GasUsed.Total.GasConsumed)
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
