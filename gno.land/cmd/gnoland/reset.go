package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"

	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

type resetCfg struct {
	baseCfg
}

func (rc *resetCfg) RegisterFlags(fs *flag.FlagSet) {}

// XXX: this is totally unsafe.
// it's only suitable for testnets.
// It could result in data loss and network disrutpion while running the node and without coordination
func newResetAllCmd(bc baseCfg) *commands.Command {
	cfg := resetCfg{
		baseCfg: bc,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "unsafe-reset-all",
			ShortUsage: "unsafe-reset-all",
			ShortHelp:  "(unsafe) remove all data, reset the node and validator to genesis state",
		},
		&cfg,
		func(_ context.Context, args []string) error {
			return execResetAll(cfg, args)
		},
	)
}

func execResetAll(rc resetCfg, args []string) (err error) {
	config := rc.tmConfig

	return resetAll(
		config.DBDir(),
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
		logger,
	)
}

// resetAll removes address book files plus all data, and resets the privValdiator data.
func resetAll(dbDir, privValKeyFile, privValStateFile string, logger log.Logger) error {
	if err := os.RemoveAll(dbDir); err == nil {
		logger.Info("Removed all blockchain history", "dir", dbDir)
	} else {
		logger.Error("Error removing all blockchain history", "dir", dbDir, "err", err)
	}

	if err := osm.EnsureDir(dbDir, 0o700); err != nil {
		logger.Error("unable to recreate dbDir", "err", err)
	}

	// recreate the dbDir since the privVal state needs to live there
	resetFilePV(privValKeyFile, privValStateFile, logger)
	return nil
}

func resetFilePV(privValKeyFile, privValStateFile string, logger log.Logger) {
	if _, err := os.Stat(privValKeyFile); err == nil {
		pv := privval.LoadFilePVEmptyState(privValKeyFile, privValStateFile)
		pv.Reset()
		logger.Info(
			"Reset private validator file to genesis state",
			"keyFile", privValKeyFile,
			"stateFile", privValStateFile,
		)
	} else {
		pv := privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		logger.Info(
			"Generated private validator file",
			"keyFile", privValKeyFile,
			"stateFile", privValStateFile,
		)
	}
}

// XXX: resetState is less risky than resetAll; however, it is still considered unsafe.
// resetState removes all databases but retains the last voting state and height of a validator.
// It is used by a validator to resync the state from other nodes, reducing the risk of double signing
// historical blocks during state syncs and the risk of a chain fork.
func newResetStateCmd(bc baseCfg) *commands.Command {
	cfg := resetCfg{
		baseCfg: bc,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "reset-state",
			ShortUsage: "reset-state",
			ShortHelp:  "reset node to genesis state, retaining validator state.",
		},
		&cfg,
		func(_ context.Context, args []string) error {
			return execResetState(cfg, args)
		},
	)
}

func execResetState(rc resetCfg, args []string) (err error) {
	config := rc.tmConfig

	return resetState(
		config.DBDir(),
		logger,
	)
}

func resetState(dbDir string, logger log.Logger) error {
	blockdb := filepath.Join(dbDir, "blockstore.db")
	state := filepath.Join(dbDir, "state.db")
	wal := filepath.Join(dbDir, "cs.wal")
	gnolang := filepath.Join(dbDir, "gnolang.db")

	removeData(blockdb)
	removeData(state)
	removeData(wal)
	removeData(gnolang)

	if err := osm.EnsureDir(dbDir, 0o700); err != nil {
		logger.Error("unable to recreate dbDir", "err", err)
	}
	return nil
}

func removeData(filepath string) {
	if osm.FileExists(filepath) {
		if err := os.RemoveAll(filepath); err == nil {
			logger.Info("Removed all", filepath)
		} else {
			logger.Error("error removing all", filepath, "err", err)
		}
	}
}
