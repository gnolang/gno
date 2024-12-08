package main

import (
	"context"
	"flag"
	"os"

	"github.com/docker/docker/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type backupCfg struct {
	rpcAddr        string
	traefikGnoFile string
	hostPWD        string

	masterBackupFile string
	snapshotsDir     string
}

func (c *backupCfg) RegisterFlags(fs *flag.FlagSet) {
	if os.Getenv("HOST_PWD") == "" {
		os.Setenv("HOST_PWD", os.Getenv("PWD"))
	}

	if os.Getenv("SNAPSHOTS_DIR") == "" {
		os.Setenv("SNAPSHOTS_DIR", "./backups/snapshots")
	}

	if os.Getenv("RPC_URL") == "" {
		os.Setenv("RPC_URL", "http://rpc.portal.gno.local:26657")
	}

	if os.Getenv("PROM_ADDR") == "" {
		os.Setenv("PROM_ADDR", ":9090")
	}

	if os.Getenv("TRAEFIK_GNO_FILE") == "" {
		os.Setenv("TRAEFIK_GNO_FILE", "./traefik/gno.yml")
	}

	if os.Getenv("MASTER_BACKUP_FILE") == "" {
		os.Setenv("MASTER_BACKUP_FILE", "./backups/backup.jsonl")
	}

	fs.StringVar(&c.rpcAddr, "rpc", os.Getenv("RPC_URL"), "tendermint rpc url")
	fs.StringVar(&c.traefikGnoFile, "traefik-gno-file", os.Getenv("TRAEFIK_GNO_FILE"), "traefik gno file")
	fs.StringVar(&c.hostPWD, "pwd", os.Getenv("HOST_PWD"), "host pwd (for docker usage)")
	fs.StringVar(&c.masterBackupFile, "master-backup-file", os.Getenv("MASTER_BACKUP_FILE"), "master txs backup file path")
	fs.StringVar(&c.snapshotsDir, "snapshots-dir", os.Getenv("SNAPSHOTS_DIR"), "snapshots directory")
}

func newBackupCmd(io commands.IO) *commands.Command {
	cfg := &backupCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "backup",
			ShortUsage: "backup [flags]",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execBackup(ctx, cfg)
		},
	)
}

func execBackup(ctx context.Context, cfg *backupCfg) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	portalLoop := &snapshotter{}

	portalLoop, err = NewSnapshotter(dockerClient, config{
		snapshotsDir:     cfg.snapshotsDir,
		masterBackupFile: cfg.masterBackupFile,
		rpcAddr:          cfg.rpcAddr,
		hostPWD:          cfg.hostPWD,
		traefikGnoFile:   cfg.traefikGnoFile,
	})
	if err != nil {
		return err
	}

	err = StartPortalLoop(ctx, portalLoop, false)
	if err != nil {
		return err
	}

	return portalLoop.backupTXs(ctx, portalLoop.url)
}
