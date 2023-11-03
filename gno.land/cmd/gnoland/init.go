package main

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	tmos "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

// Display a node's persistent peer ID to the standard output.
func newInitCmd(bc baseCfg) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init",
			ShortHelp:  "initialize gnoland node",
		},
		nil,
		func(_ context.Context, args []string) error {
			return execInit(bc)
		},
	)
	return cmd
}

func execInit(bc baseCfg) error {
	config := bc.tmConfig
	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	if tmos.FileExists(privValKeyFile) {
		logger.Info("Found private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		logger.Info("Generated private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if tmos.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	return nil
}
