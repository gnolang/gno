package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/std"

	// XXX better way?
	_ "github.com/gnolang/gno/pkgs/sdk/auth"
	_ "github.com/gnolang/gno/pkgs/sdk/bank"
	_ "github.com/gnolang/gno/pkgs/sdk/vm"
)

type txExportOptions struct {
	Remote      string `flag:"remote" help:"Remote RPC addr:port"`
	StartHeight int64  `flag:"start" help:"Start height"`
	EndHeight   int64  `flag:"end" help:"End height (optional)"`
	OutFile     string `flag:"out" help:"Output file path"`
	Quiet       bool   `flag:"quiet" help:"Quiet mode"`
	Follow      bool   `flag:"follow" help:"Keep attached and follow new events"`
}

var defaultTxExportOptions = txExportOptions{
	Remote:      "localhost:26657",
	StartHeight: 1,
	EndHeight:   0,
	OutFile:     "txexport.log",
	Quiet:       false,
	Follow:      false,
}

func txExportApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(txExportOptions)
	c := client.NewHTTP(opts.Remote, "/websocket")
	status, err := c.Status()
	if err != nil {
		panic(err)
	}
	last := int64(0)
	if opts.EndHeight == 0 {
		last = status.SyncInfo.LatestBlockHeight
	} else {
		last = opts.EndHeight
	}
	var out io.Writer
	switch opts.OutFile {
	case "-", "STDOUT":
		out = os.Stdout
	default:
		out, err = os.OpenFile(opts.OutFile, os.O_RDWR|os.O_CREATE, 0o755)
		if err != nil {
			return err
		}
	}

	for height := opts.StartHeight; ; height++ {
		if !opts.Follow && height >= last {
			break
		}

	getBlock:
		block, err := c.Block(&height)
		if err != nil {
			if opts.Follow && strings.Contains(err.Error(), "") {
				time.Sleep(time.Second)
				goto getBlock
			}
			panic(err)
		}
		txs := block.Block.Data.Txs
		if len(txs) == 0 {
			continue
		}
		_, err = c.BlockResults(&height)
		if err != nil {
			if opts.Follow && strings.Contains(err.Error(), "") {
				time.Sleep(time.Second)
				goto getBlock
			}
			panic(err)
		}
		for i := 0; i < len(txs); i++ {
			// need to include error'd txs, to keep sequence alignment.
			//if bres.Results.DeliverTxs[i].Error != nil {
			//	continue
			//}
			tx := txs[i]
			stdtx := std.Tx{}
			amino.MustUnmarshal(tx, &stdtx)
			bz := amino.MustMarshalJSON(stdtx)
			fmt.Fprintln(out, string(bz))
		}
		if !opts.Quiet {
			log.Printf("h=%d/%d (txs=%d)", height, last, len(txs))
		}
	}
	return nil
}
