package main

import (
	"io/ioutil"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/std"

	// XXX better way?
	_ "github.com/gnolang/gno/pkgs/sdk/auth"
	_ "github.com/gnolang/gno/pkgs/sdk/bank"
	_ "github.com/gnolang/gno/pkgs/sdk/vm"
)

type txImportOptions struct {
	Remote string `flag:"remote" help:"Remote RPC addr:port"`
	InFile string `flag:"in" help:"Input file path"`
}

var defaultTxImportOptions = txImportOptions{
	Remote: "gno.land:36657",
	InFile: "txexport.log",
}

func txImportApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(txImportOptions)
	c := client.NewHTTP(opts.Remote, "/websocket")
	filebz, err := ioutil.ReadFile(opts.InFile)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(filebz)), "\n")
	for i, line := range lines {
		if len(line) == 0 {
			panic(i)
		}
		var tx std.Tx
		amino.MustUnmarshalJSON([]byte(line), &tx)
		txbz := amino.MustMarshal(tx)
		_, err := c.BroadcastTxSync(txbz)
		if err != nil {
			return errors.Wrap(err, "broadcasting tx %d", i)
		}
	}
	return nil
}
