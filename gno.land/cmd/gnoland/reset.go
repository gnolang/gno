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
			ShortHelp:  "(unsafe) Remove all the data and WAL, reset this node's validator to genesis state",
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

// resetState removes address book files plus all databases.
func resetState(dbDir string, logger log.Logger) error {
	blockdb := filepath.Join(dbDir, "blockstore.db")
	state := filepath.Join(dbDir, "state.db")
	wal := filepath.Join(dbDir, "cs.wal")
	gnolang := filepath.Join(dbDir, "gnolang.db")

	if osm.FileExists(blockdb) {
		if err := os.RemoveAll(blockdb); err == nil {
			logger.Info("Removed all blockstore.db", "dir", blockdb)
		} else {
			logger.Error("error removing all blockstore.db", "dir", blockdb, "err", err)
		}
	}

	if osm.FileExists(state) {
		if err := os.RemoveAll(state); err == nil {
			logger.Info("Removed all state.db", "dir", state)
		} else {
			logger.Error("error removing all state.db", "dir", state, "err", err)
		}
	}

	if osm.FileExists(wal) {
		if err := os.RemoveAll(wal); err == nil {
			logger.Info("Removed all cs.wal", "dir", wal)
		} else {
			logger.Error("error removing all cs.wal", "dir", wal, "err", err)
		}
	}

	if osm.FileExists(gnolang) {
		if err := os.RemoveAll(gnolang); err == nil {
			logger.Info("Removed all gnolang.db", "dir", gnolang)
		} else {
			logger.Error("error removing all gnolang.db", "dir", gnolang, "err", err)
		}
	}

	if err := osm.EnsureDir(dbDir, 0o700); err != nil {
		logger.Error("unable to recreate dbDir", "err", err)
	}
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
