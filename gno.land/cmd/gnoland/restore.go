package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/backup/v1"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type restoreCfg struct {
	nodeCfg
	backupDir string
	endHeight int64
}

func newRestoreCmd(io commands.IO) *commands.Command {
	cfg := &restoreCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "restore",
			ShortUsage: "restore [flags]",
			ShortHelp:  "restore the Gnoland blockchain node",
			LongHelp:   "Restores the Gnoland blockchain node, with accompanying setup",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execRestore(ctx, cfg, io)
		},
	)
}

func (c *restoreCfg) RegisterFlags(fs *flag.FlagSet) {
	c.nodeCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.backupDir,
		"backup-dir",
		"blocks-backup",
		"directory where the backup files are",
	)

	fs.Int64Var(
		&c.endHeight,
		"end-height",
		0,
		"height at which the restore process should stop",
	)
}

func execRestore(ctx context.Context, c *restoreCfg, io commands.IO) error {
	gnoNode, err := createNode(&c.nodeCfg, io)
	if err != nil {
		return err
	}

	// need block n+1 to commit block n
	endHeight := c.endHeight
	if endHeight != 0 {
		endHeight += 1
	}

	startHeight := gnoNode.BlockStore().Height() + 1
	if c.endHeight != 0 && c.endHeight < startHeight {
		return fmt.Errorf("invalid input: requested end height (#%d) is smaller than next chain height (#%d)", c.endHeight, startHeight)
	}

	return backup.WithReader(c.backupDir, startHeight, endHeight, func(reader backup.Reader) error {
		return gnoNode.Restore(ctx, reader)
	})
}
