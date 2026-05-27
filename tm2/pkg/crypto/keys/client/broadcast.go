package client

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

type BroadcastCfg struct {
	RootCfg *BaseCfg

	DryRun bool

	// internal
	tx *std.Tx
	// Set by SignAndBroadcastHandler, similar to DryRun.
	// If true, simulation is attempted but not printed;
	// the result is only returned in case of an error.
	testSimulate bool
	// max gas limit to use for simulation (optional).
	simulateMaxGas int64
	GasFeeMargin   uint64
}

const simulationMaxGasFallback = int64(math.MaxInt64)

func NewBroadcastCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &BroadcastCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "broadcast",
			ShortUsage: "broadcast [flags] <file-name>",
			ShortHelp:  "broadcasts a signed document",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execBroadcast(cfg, args, io)
		},
	)
}

func (c *BroadcastCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.DryRun,
		"dry-run",
		false,
		"perform a dry-run broadcast",
	)
}

func execBroadcast(cfg *BroadcastCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}
	filename := args[0]

	jsonbz, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "reading tx document file "+filename)
	}
	var tx std.Tx
	err = amino.UnmarshalJSON(jsonbz, &tx)
	if err != nil {
		return errors.Wrap(err, "unmarshaling tx json bytes")
	}
	cfg.tx = &tx

	res, err := BroadcastHandler(cfg)
	if err != nil {
		return err
	}

	if res.CheckTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", res, res.CheckTx.Log)
	} else if res.DeliverTx.IsErr() {
		io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(res.Hash))
		return errors.New("transaction failed %#v\nlog %s", res, res.DeliverTx.Log)
	} else {
		if cfg.RootCfg.OnTxSuccess != nil {
			cfg.RootCfg.OnTxSuccess(tx, res)
		} else {
			io.Println(string(res.DeliverTx.Data))
			io.Println("OK!")
			io.Println("GAS WANTED:", res.DeliverTx.GasWanted)
			io.Println("GAS USED:  ", res.DeliverTx.GasUsed)
			io.Println("HEIGHT:    ", res.Height)
			io.Println("EVENTS:    ", string(res.DeliverTx.EncodeEvents()))
			io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(res.Hash))
		}
	}
	return nil
}

func BroadcastHandler(cfg *BroadcastCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	if cfg.tx == nil {
		return nil, errors.New("invalid tx")
	}

	remote := cfg.RootCfg.Remote
	if remote == "" {
		return nil, errors.New("missing remote url")
	}

	bz, err := amino.Marshal(cfg.tx)
	if err != nil {
		return nil, errors.Wrap(err, "remarshaling tx binary bytes")
	}

	// NewHTTPClient applies a 60s request timeout to every call, so
	// context.Background() (no deadline) is safe to use here.
	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, err
	}

	// Both for DryRun and testSimulate, we perform simulation.
	// However, DryRun always returns here, while in case of success
	// testSimulate continues onto broadcasting the transaction.
	if cfg.DryRun || cfg.testSimulate {
		simBz, rewritten, err := buildSimulationTxBytes(cfg.tx, bz, cfg.simulateMaxGas)
		if err != nil {
			return nil, err
		}

		originalGasWanted := cfg.tx.Fee.GasWanted
		res, err := SimulateTx(cli, simBz)
		if rewritten && res != nil {
			res.DeliverTx.GasWanted = originalGasWanted
			if originalGasWanted > 0 && res.DeliverTx.Error == nil && res.DeliverTx.GasUsed > originalGasWanted {
				log := store.OutOfGasLog(res.DeliverTx.GasUsed, originalGasWanted, cfg.simulateMaxGas, "simulation", false)
				res.DeliverTx.Error = abci.ABCIErrorOrStringError(std.ErrOutOfGas(log))
				res.DeliverTx.Log = log
			}
		}
		if res != nil {
			hasError := err != nil || res.CheckTx.IsErr() || res.DeliverTx.IsErr()
			if cfg.DryRun && !hasError {
				err = estimateGasFee(cli, res, cfg.GasFeeMargin)
				return res, err
			}
			appendSuggestedGasWanted(res)
			if hasError {
				return res, err
			}
		} else if err != nil {
			return nil, err
		}
	}

	bres, err := cli.BroadcastTxCommit(context.Background(), bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	return bres, nil
}

// buildSimulationTxBytes returns tx bytes to use for simulation, overriding
// GasWanted to consensus maxGas. If maxGas is -1 (chain has no gas limit) it
// falls back to MaxInt64. If maxGas is 0 (unknown, e.g. fetch failed) the
// original bytes are returned unchanged. It also returns whether tx bytes were
// rewritten.
func buildSimulationTxBytes(tx *std.Tx, txBytes []byte, maxGas int64) ([]byte, bool, error) {
	switch maxGas {
	case 0:
		return txBytes, false, nil
	case -1:
		maxGas = simulationMaxGasFallback
	}
	if tx.Fee.GasWanted >= maxGas {
		return txBytes, false, nil
	}

	simTx := *tx
	simTx.Fee.GasWanted = maxGas
	simBz, err := amino.Marshal(&simTx)
	if err != nil {
		return nil, false, errors.Wrap(err, "remarshaling tx binary bytes for simulation")
	}

	return simBz, true, nil
}

func suggestedGasWanted(gasUsed int64) int64 {
	margin := gasUsed / 20
	if gasUsed%20 != 0 {
		margin++
	}
	return overflow.Addp(gasUsed, margin)
}

func appendSuggestedGasWanted(bres *ctypes.ResultBroadcastTxCommit) {
	suggested := suggestedGasWanted(bres.DeliverTx.GasUsed)
	msg := fmt.Sprintf("suggested gas-wanted (gas used + 5%%): %d", suggested)
	if bres.DeliverTx.Info == "" {
		bres.DeliverTx.Info = msg
	} else {
		bres.DeliverTx.Info = bres.DeliverTx.Info + ", " + msg
	}
}

func estimateGasFee(cli client.ABCIClient, bres *ctypes.ResultBroadcastTxCommit, gasFeeMargin uint64) error {
	gasUsed := bres.DeliverTx.GasUsed
	suggested := suggestedGasWanted(gasUsed)

	gp := std.GasPrice{}
	qres, err := cli.ABCIQuery(context.Background(), "auth/gasprice", []byte{})
	if err != nil {
		return errors.Wrap(err, "query gas price")
	}
	err = amino.UnmarshalJSON(qres.Response.Data, &gp)
	if err != nil {
		return errors.Wrap(err, "unmarshaling query gas price result")
	}

	var s string
	if gp.Gas == 0 {
		s = fmt.Sprintf("estimated gas usage: %d (suggested, with 5%% margin: %d)\n", gasUsed, suggested)
	} else {
		fee := gasUsed/gp.Gas + 1
		fee = overflow.Mulp(fee, gp.Price.Amount)
		// fee buffer to cover the sudden change of gas price
		feeBuffer := overflow.Mulp(fee, int64(gasFeeMargin)) / 100
		fee = overflow.Addp(fee, feeBuffer)
		s = fmt.Sprintf("estimated gas usage: %d (suggested, with 5%% margin: %d), gas fee: %d%s, current gas price: %s\n", gasUsed, suggested, fee, gp.Price.Denom, gp.String())
	}
	if bres.DeliverTx.Info == "" {
		bres.DeliverTx.Info = s
	} else {
		bres.DeliverTx.Info = bres.DeliverTx.Info + ", " + s
	}
	return nil
}

func SimulateTx(cli client.ABCIClient, tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	bres, err := cli.ABCIQuery(context.Background(), ".app/simulate", tx)
	if err != nil {
		return nil, errors.Wrap(err, "simulate tx")
	}

	var result abci.ResponseDeliverTx
	err = amino.Unmarshal(bres.Response.Value, &result)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling simulate result")
	}

	return &ctypes.ResultBroadcastTxCommit{
		DeliverTx: result,
	}, nil
}
