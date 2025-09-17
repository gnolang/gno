// Dedicated to my love, Lexi.
package keyscli

import (
	"encoding/base64"
	"slices"

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
	if bytesDelta, coinsDelta := GetStorageInfo(res.DeliverTx.Events); bytesDelta != 0 {
		io.Printfln("STORAGE DELTA:  %d bytes", bytesDelta)
		if coinsDelta.IsAllPositive() || coinsDelta.IsZero() {
			io.Println("STORAGE FEE:   ", coinsDelta)
		} else {
			// NOTE: there is edge cases where coinsDelta can be a mixture of positive and negative coins.
			// For example if a tx contains a storage cost param change message sandwiched by storage movement messages.
			// These will fall in this case and print confusing information but it's so rare that we don't
			// really care about this possibility here.
			io.Println("STORAGE REFUND:", negateCoins(coinsDelta))
		}
		io.Printfln("TOTAL TX COST:  %s", combineCoins(std.Coins{tx.Fee.GasFee}, coinsDelta))
	}
	io.Println("EVENTS:    ", string(res.DeliverTx.EncodeEvents()))
	io.Println("INFO:      ", res.DeliverTx.Info)
	io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(res.Hash))
}

// GetStorageInfo searches events for StorageDepositEvent or StorageUnlockEvent and returns the bytes delta and coins delta.
func GetStorageInfo(events []abci.Event) (int64, std.Coins) {
	var (
		bytesDelta int64
		coinsDelta std.Coins
	)

	for _, event := range events {
		switch storageEvent := event.(type) {
		case gnostd.StorageDepositEvent:
			bytesDelta += storageEvent.BytesDelta
			coinsDelta = combineCoins(coinsDelta, std.Coins{storageEvent.FeeDelta})
		case gnostd.StorageUnlockEvent:
			bytesDelta += storageEvent.BytesDelta
			if !storageEvent.RefundWithheld {
				coinsDelta = combineCoins(coinsDelta, negateCoins(std.Coins{storageEvent.FeeRefund}))
			}
		}
	}

	return bytesDelta, coinsDelta
}

func combineCoins(bags ...std.Coins) std.Coins {
	res := std.Coins{}
	for _, bag := range bags {
		for _, coin := range bag {
			if coin.Amount == 0 {
				continue
			}
			indexInRes := slices.IndexFunc(res, func(resElem std.Coin) bool { return resElem.Denom == coin.Denom })
			if indexInRes == -1 {
				res = append(res, coin)
				continue
			}
			res[indexInRes].Amount += coin.Amount
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res.Sort()
}

func negateCoins(coins std.Coins) std.Coins {
	res := make(std.Coins, len(coins))
	for i, coin := range coins {
		res[i] = std.Coin{Denom: coin.Denom, Amount: -coin.Amount}
	}
	return res
}
