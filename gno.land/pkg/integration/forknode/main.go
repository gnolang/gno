package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// isAllZero checks if all elements in the [64]byte array are zero.
func isAllZero(arr [64]byte) bool {
	for _, v := range arr {
		if v != 0 {
			return false
		}
	}
	return true
}

func ForkableNode(cfg *integration.ForkConfig) error {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var data db.DB
	var err error
	if cfg.DBDir == "" {
		data = memdb.NewMemDB()
	} else {
		data, err = goleveldb.NewGoLevelDB("testdb", cfg.DBDir)
		if err != nil {
			return fmt.Errorf("unable to init database in %q: %w", cfg.DBDir, err)
		}
	}

	nodecfg := integration.TestingMinimalNodeConfig(cfg.RootDir)

	if len(cfg.PrivValidator) > 0 && !isAllZero(cfg.PrivValidator) {
		nodecfg.PrivValidator = bft.NewMockPVWithParams(cfg.PrivValidator, false, false)
		pv := nodecfg.PrivValidator.GetPubKey()
		nodecfg.Genesis.Validators = []bft.GenesisValidator{
			{
				Address: pv.Address(),
				PubKey:  pv,
				Power:   10,
				Name:    "self",
			},
		}

	}

	nodecfg.DB = data
	nodecfg.TMConfig.DBPath = cfg.DBDir
	nodecfg.TMConfig = cfg.TMConfig
	nodecfg.Genesis = cfg.Genesis.ToGenesisDoc()

	node, err := gnoland.NewInMemoryNode(logger, nodecfg)
	if err != nil {
		return fmt.Errorf("failed to create new in-memory node: %w", err)
	}

	err = node.Start()
	if err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	ourAddress := nodecfg.PrivValidator.GetPubKey().Address()
	isValidator := slices.ContainsFunc(nodecfg.Genesis.Validators, func(val bft.GenesisValidator) bool {
		return val.Address == ourAddress
	})

	// Wait for first block if we are a validator.
	// If we are not a validator, we don't produce blocks, so node.Ready() hangs.
	if isValidator {
		select {
		case <-node.Ready():
			fmt.Printf("READY:%s\n", node.Config().RPC.ListenAddress)
		case <-time.After(time.Second * 10):
			return fmt.Errorf("timeout while waiting for the node to start")
		}
	} else {
		fmt.Printf("READY:%s\n", node.Config().RPC.ListenAddress)
	}

	// Keep the function running indefinitely if no errors occur
	select {}
}

func main() {
	// Read the configuration from standard input
	configData, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	// Unmarshal the JSON configuration
	var cfg integration.ForkConfig
	err = json.Unmarshal(configData, &cfg)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error unmarshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Call the ForkableNode function with the parsed configuration
	if err := ForkableNode(&cfg); err != nil {
		fmt.Fprintf(os.Stdout, "Error running ForkableNode: %v\n", err)
		os.Exit(1)
	}
}
