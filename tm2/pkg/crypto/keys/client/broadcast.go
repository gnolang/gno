package client

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/std"
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
}

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

func BroadcastHandler(cfg *BroadcastCfg) (*mempool.ResultBroadcastTxCommit, error) {
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

	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, err
	}

	// Both for DryRun and testSimulate, we perform simulation.
	// However, DryRun always returns here, while in case of success
	// testSimulate continues onto broadcasting the transaction.
	if cfg.DryRun || cfg.testSimulate {
		res, err := SimulateTx(cli, bz)
		hasError := err != nil || res.CheckTx.IsErr() || res.DeliverTx.IsErr()
		if hasError {
			return res, err
		}
		if cfg.DryRun { // we estmate the gas fee in dry run
			err = estimateGasFee(cli, res)
			return res, err
		}
	}

	bres, err := cli.BroadcastTxCommit(context.Background(), bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	return bres, nil
}

func estimateGasFee(cli client.ABCIClient, bres *mempool.ResultBroadcastTxCommit) error {
	gp := std.GasPrice{}
	qres, err := cli.ABCIQuery(context.Background(), "auth/gasprice", []byte{})
	if err != nil {
		return errors.Wrap(err, "query gas price")
	}
	err = amino.UnmarshalJSON(qres.Response.Data, &gp)
	if err != nil {
		return errors.Wrap(err, "unmarshaling query gas price result")
	}

	if gp.Gas == 0 {
		return nil
	}

	fee := bres.DeliverTx.GasUsed/gp.Gas + 1
	fee = overflow.Mulp(fee, gp.Price.Amount)
	// 5% fee buffer to cover the suden change of gas price
	feeBuffer := overflow.Mulp(fee, 5) / 100
	fee = overflow.Addp(fee, feeBuffer)
	s := fmt.Sprintf("estimated gas usage: %d, gas fee: %d%s, current gas price: %s\n", bres.DeliverTx.GasUsed, fee, gp.Price.Denom, gp.String())
	bres.DeliverTx.Info = s
	return nil
}

func SimulateTx(cli client.ABCIClient, tx []byte) (*mempool.ResultBroadcastTxCommit, error) {
	bres, err := cli.ABCIQuery(context.Background(), ".app/simulate", tx)
	if err != nil {
		return nil, errors.Wrap(err, "simulate tx")
	}

	var result abci.ResponseDeliverTx
	err = amino.Unmarshal(bres.Response.Value, &result)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling simulate result")
	}

	return &mempool.ResultBroadcastTxCommit{
		DeliverTx: result,
	}, nil
}
