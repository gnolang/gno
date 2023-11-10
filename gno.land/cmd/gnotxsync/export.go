package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"

	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	_ "github.com/gnolang/gno/tm2/pkg/sdk/auth" // XXX better way?
	_ "github.com/gnolang/gno/tm2/pkg/sdk/bank"
)

type exportCfg struct {
	rootCfg *config

	startHeight int64
	tailHeight  int64
	endHeight   int64
	outFile     string
	quiet       bool
	follow      bool
}

func newExportCommand(rootCfg *config) *commands.Command {
	cfg := &exportCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "export",
			ShortUsage: "export [flags] <file>",
			ShortHelp:  "Export transactions to file",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execExport(cfg)
		},
	)
}

func (c *exportCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.Int64Var(&c.startHeight, "start", 1, "start height")
	fs.Int64Var(&c.tailHeight, "tail", 0, "start at LAST - N")
	fs.Int64Var(&c.endHeight, "end", 0, "end height (optional)")
	fs.StringVar(&c.outFile, "out", defaultFilePath, "output file path")
	fs.BoolVar(&c.quiet, "quiet", false, "omit console output during execution")
	fs.BoolVar(&c.follow, "follow", false, "keep attached and follow new events")
}

func execExport(c *exportCfg) error {
	node := client.NewHTTP(c.rootCfg.remote, "/websocket")

	status, err := node.Status()
	if err != nil {
		return fmt.Errorf("unable to fetch node status, %w", err)
	}

	var (
		start = c.startHeight
		end   = c.endHeight
		tail  = c.tailHeight
	)

	if end == 0 { // take last block height
		end = status.SyncInfo.LatestBlockHeight
	}
	if tail > 0 {
		start = end - tail
	}

	var out io.Writer
	switch c.outFile {
	case "-", "STDOUT":
		out = os.Stdout
	default:
		out, err = os.OpenFile(c.outFile, os.O_RDWR|os.O_CREATE, 0o755)
		if err != nil {
			return err
		}
	}

	for height := start; ; height++ {
		if !c.follow && height >= end {
			break
		}

	getBlock:
		block, err := node.Block(&height)
		if err != nil {
			if c.follow && strings.Contains(err.Error(), "") {
				time.Sleep(time.Second)

				goto getBlock
			}

			return fmt.Errorf("encountered error while fetching block, %w", err)
		}

		txs := block.Block.Data.Txs
		if len(txs) == 0 {
			continue
		}

		_, err = node.BlockResults(&height)
		if err != nil {
			if c.follow && strings.Contains(err.Error(), "") {
				time.Sleep(time.Second)

				goto getBlock
			}

			return fmt.Errorf("encountered error while fetching block results, %w", err)
		}

		for i := 0; i < len(txs); i++ {
			tx := txs[i]
			stdtx := std.Tx{}

			amino.MustUnmarshal(tx, &stdtx)

			bz := amino.MustMarshalJSON(stdtx)

			_, _ = fmt.Fprintln(out, string(bz))
		}

		if !c.quiet {
			log.Printf("h=%d/%d (txs=%d)", height, end, len(txs))
		}
	}

	return nil
}
