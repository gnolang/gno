package main

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

// Display a node's persistent peer ID to the standard output.
func newNodeIDCmd(bc baseCfg) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "node",
			ShortUsage: "node",
			ShortHelp:  "display the node id for configuring persistent peers",
		},
		nil,
		func(_ context.Context, args []string) error {
			return execNodeID(bc)
		},
	)
	return cmd
}

func execNodeID(bc baseCfg) error {
	config := bc.tmConfig
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}

	fmt.Printf("NodeID: %v\n", nodeKey.ID())
	return nil
}
