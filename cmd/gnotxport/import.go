package main

import (
	"fmt"
	"os"
	"strings"
	"time"

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
	Remote: "localhost:26657",
	InFile: "txexport.log",
}

func txImportApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(txImportOptions)
	c := client.NewHTTP(opts.Remote, "/websocket")
	filebz, err := os.ReadFile(opts.InFile)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(filebz)), "\n")
	for i, line := range lines {
		print(".")
		// time.Sleep(10 * time.Second)
		if len(line) == 0 {
			panic(i)
		}
		var tx std.Tx
		amino.MustUnmarshalJSON([]byte(line), &tx)
		txbz := amino.MustMarshal(tx)
		res, err := c.BroadcastTxSync(txbz)
		if err != nil || res.Error != nil {
			print("!")
			// wait for next block and try again.
			// TODO: actually wait 1 block instead of fudging it.
			time.Sleep(20 * time.Second)
			res, err := c.BroadcastTxSync(txbz)
			if err != nil || res.Error != nil {
				if err != nil {
					fmt.Println("SECOND ERROR", err)
				} else {
					fmt.Println("SECOND ERROR!", res.Error)
				}
				fmt.Println(line)
				return errors.Wrap(err, "broadcasting tx %d", i)
			}
		}
	}
	return nil
}
