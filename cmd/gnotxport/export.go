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
	TailHeight int64 `flag:"tail" help:"Start at LAST - N"`
	EndHeight   int64  `flag:"end" help:"End height (optional)"`
	OutFile     string `flag:"out" help:"Output file path"`
	Quiet       bool   `flag:"quiet" help:"Quiet mode"`
	Follow      bool   `flag:"follow" help:"Keep attached and follow new events"`
}

var defaultTxExportOptions = txExportOptions{
	Remote:      "localhost:26657",
	StartHeight: 1,
	EndHeight:   0,
	TailHeight: 0,
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
	start := opts.StartHeight
	end := opts.EndHeight
	tail := opts.TailHeight
	if end == 0 { // take last block height
		end = status.SyncInfo.LatestBlockHeight
	}
	if tail > 0 {
		start = end - tail
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

	for height := start; ; height++ {
		if !opts.Follow && height >= end {
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
			log.Printf("h=%d/%d (txs=%d)", height, end, len(txs))
		}
	}
	return nil
}
