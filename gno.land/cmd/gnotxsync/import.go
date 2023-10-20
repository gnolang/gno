package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"

	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	_ "github.com/gnolang/gno/tm2/pkg/sdk/auth" // XXX better way?
	_ "github.com/gnolang/gno/tm2/pkg/sdk/bank"
)

type importCfg struct {
	rootCfg *config

	inFile string
}

func newImportCommand(rootCfg *config) *commands.Command {
	cfg := &importCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "import",
			ShortUsage: "import [flags] <file>",
			ShortHelp:  "Import transactions from file",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execImport(ctx, cfg)
		},
	)
}

func (c *importCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.inFile, "in", defaultFilePath, "input file path")
}

func execImport(ctx context.Context, c *importCfg) error {
	// Initial validation
	if len(c.inFile) == 0 {
		return errors.New("input file path not specified")
	}

	// Read the input file
	file, err := os.Open(c.inFile)
	if err != nil {
		return fmt.Errorf("unable to open input file, %w", err)
	}

	defer file.Close()

	// Start the WS connection to the node
	node := client.NewHTTP(c.rootCfg.remote, "/websocket")

	index := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			// Stop signal received while parsing
			// the import file
			return nil
		default:
			print(".")

			line := scanner.Text()
			if len(line) == 0 {
				return fmt.Errorf("empty line encountered at %d", index)
			}

			var tx std.Tx
			amino.MustUnmarshalJSON([]byte(line), &tx)
			txbz := amino.MustMarshal(tx)

			res, err := node.BroadcastTxSync(txbz)

			if err != nil || res.Error != nil {
				print("!")
				// wait for next block and try again.
				// TODO: actually wait 1 block instead of fudging it.
				time.Sleep(20 * time.Second)

				res, err := node.BroadcastTxSync(txbz)
				if err != nil || res.Error != nil {
					if err != nil {
						fmt.Println("SECOND ERROR", err)
					} else {
						fmt.Println("SECOND ERROR!", res.Error)
					}

					fmt.Println(line)

					return errors.Wrap(err, "broadcasting tx %d", index)
				}
			}

			index++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error encountered while reading file, %w", err)
	}

	return nil
}
