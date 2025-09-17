// Dedicated to my love, Lexi.
package keyscli

import (
	"encoding/base64"

	gnostd "github.com/gnolang/gno/gnovm/stdlibs/std"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
)

func NewRootCmd(io commands.IO, base client.BaseOptions) *commands.Command {
	cfg := &client.BaseCfg{
		BaseOptions: base,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "gno.land keychain & client",
			Options: []ff.Option{
				ff.WithConfigFileFlag("config"),
				ff.WithConfigFileParser(fftoml.Parser),
			},
		},
		cfg,
		commands.HelpExec,
	)

	// OnTxSuccess is only used by NewBroadcastCmd
	cfg.OnTxSuccess = func(tx std.Tx, res *ctypes.ResultBroadcastTxCommit) {
		PrintTxInfo(tx, res, io)
	}
	cmd.AddSubCommands(
		client.NewAddCmd(cfg, io),
		client.NewDeleteCmd(cfg, io),
		client.NewRotateCmd(cfg, io),
		client.NewGenerateCmd(cfg, io),
		client.NewExportCmd(cfg, io),
		client.NewImportCmd(cfg, io),
		client.NewListCmd(cfg, io),
		client.NewSignCmd(cfg, io),
		client.NewVerifyCmd(cfg, io),
		client.NewQueryCmd(cfg, io),
		client.NewBroadcastCmd(cfg, io),
		client.NewMultisignCmd(cfg, io),

		// Custom MakeTX command
		NewMakeTxCmd(cfg, io),
	)

	return cmd
}

// PrintTxInfo prints the transaction result to io. If the events has storage deposit
// info then also print it with the total transaction cost.
func PrintTxInfo(tx std.Tx, res *ctypes.ResultBroadcastTxCommit, io commands.IO) {
	io.Println(string(res.DeliverTx.Data))
	io.Println("OK!")
	io.Println("GAS WANTED:", res.DeliverTx.GasWanted)
	io.Println("GAS USED:  ", res.DeliverTx.GasUsed)
	io.Println("HEIGHT:    ", res.Height)
	if delta, storageFee, ok := GetStorageInfo(res.DeliverTx.Events); ok {
		io.Printfln("STORAGE DELTA:  %d bytes", delta)
		total := tx.Fee.GasFee.Amount

		if storageFee.Amount >= 0 {
			io.Println("STORAGE FEE:   ", storageFee)
		} else {
			refund := storageFee
			refund.Amount = -refund.Amount
			io.Println("STORAGE REFUND:", refund)
		}
		if tx.Fee.GasFee.Denom == storageFee.Denom {
			total := tx.Fee.GasFee.Amount + storageFee.Amount
			io.Printfln("TOTAL TX COST:  %d%v", total, tx.Fee.GasFee.Denom)
		}

		io.Printfln("TOTAL TX COST:  %d%v", total, tx.Fee.GasFee.Denom)
	}
	io.Println("EVENTS:    ", string(res.DeliverTx.EncodeEvents()))
	io.Println("INFO:      ", res.DeliverTx.Info)
	io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(res.Hash))
}

// GetStorageInfo searches events for StorageDepositEvent or StorageUnlockEvent and returns the bytes delta and fee.
// If this is "unlock", then bytes delta and fee are negative.
// The third return is true if found, else false.
func GetStorageInfo(events []abci.Event) (int64, std.Coin, bool) {
	for _, event := range events {
		switch storageEvent := event.(type) {
		case gnostd.StorageDepositEvent:
			return storageEvent.BytesDelta, storageEvent.FeeDelta, true
		case gnostd.StorageUnlockEvent:
			fee := storageEvent.FeeRefund
			fee.Amount *= -1
			// If true it means the refund was withheld
			// due to token lock, so the refund visible to user is 0
			if storageEvent.RefundWithheld {
				fee.Amount = 0
			}
			// For unlock, BytesDelta is negative
			return storageEvent.BytesDelta, fee, true
		}
	}

	return 0, std.Coin{}, false
}
