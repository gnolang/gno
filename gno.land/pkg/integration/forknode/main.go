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
)

func ForkableNode(cfg *integration.ForkConfig) error {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	nodecfg := integration.TestingMinimalNodeConfig(cfg.RootDir)
	pv := nodecfg.PrivValidator.GetPubKey()
	nodecfg.TMConfig = cfg.TMConfig
	nodecfg.Genesis = cfg.Genesis.ToGenesisDoc()
	nodecfg.Genesis.Validators = []bft.GenesisValidator{
		{
			Address: pv.Address(),
			PubKey:  pv,
			Power:   10,
			Name:    "self",
		},
	}

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
