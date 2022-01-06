package client

import (
	"io/ioutil"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/std"
)

type BroadcastOptions struct {
	BaseOptions
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
	remote := opts.Remote
	if remote == "" || remote == "y" {
		return errors.New("missing remote url")
	}

	jsonbz, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "reading tx document file "+filename)
	}
	var tx std.Tx
	err = amino.UnmarshalJSON(jsonbz, &tx)
	if err != nil {
		return errors.Wrap(err, "unmarshaling tx json bytes")
	}
	bz, err := amino.Marshal(tx)
	if err != nil {
		return errors.Wrap(err, "remarshaling tx binary bytes")
	}

	cli := client.NewHTTP(remote, "/websocket")

	bres, err := cli.BroadcastTxCommit(bz)
	if err != nil {
		return errors.Wrap(err, "broadcasting bytes")
	}
	if bres.CheckTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", bres, bres.CheckTx.Log)
	} else if bres.DeliverTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", bres, bres.DeliverTx.Log)
	} else {
		cmd.Println(string(bres.DeliverTx.Data))
		cmd.Println("OK!")
		cmd.Println("GAS WANTED:", bres.DeliverTx.GasWanted)
		cmd.Println("GAS USED:  ", bres.DeliverTx.GasUsed)
	}
	return nil
}
