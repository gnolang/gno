package client

import (
	"os"

	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/std"
)

type BroadcastOptions struct {
	BaseOptions

	// internal
	Tx     *std.Tx `flag:"-"`
	DryRun bool    `flag:"dry-run"`
}

var DefaultBroadcastOptions = BroadcastOptions{
	BaseOptions: DefaultBaseOptions,
}

func broadcastApp(cmd *command.Command, args []string, iopts interface{}) error {
	var opts BroadcastOptions = iopts.(BroadcastOptions)

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: broadcast <filename>")
		return errors.New("invalid args")
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
	opts.Tx = &tx

	res, err := BroadcastHandler(opts)
	if err != nil {
		return err
	}

	if res.CheckTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", res, res.CheckTx.Log)
	} else if res.DeliverTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", res, res.DeliverTx.Log)
	} else {
		cmd.Println(string(res.DeliverTx.Data))
		cmd.Println("OK!")
		cmd.Println("GAS WANTED:", res.DeliverTx.GasWanted)
		cmd.Println("GAS USED:  ", res.DeliverTx.GasUsed)
	}
	return nil
}

func BroadcastHandler(opts BroadcastOptions) (*ctypes.ResultBroadcastTxCommit, error) {
	if opts.Tx == nil {
		return nil, errors.New("invalid tx")
	}

	remote := opts.Remote
	if remote == "" || remote == "y" {
		return nil, errors.New("missing remote url")
	}

	bz, err := amino.Marshal(opts.Tx)
	if err != nil {
		return nil, errors.Wrap(err, "remarshaling tx binary bytes")
	}

	cli := client.NewHTTP(remote, "/websocket")

	if opts.DryRun {
		return SimulateTx(cli, bz)
	}

	bres, err := cli.BroadcastTxCommit(bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	return bres, nil
}

func SimulateTx(cli client.ABCIClient, tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	bres, err := cli.ABCIQuery(".app/simulate", tx)
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
