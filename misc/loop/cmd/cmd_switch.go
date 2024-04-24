package main

import (
	"context"
	"flag"
	"os"

	"github.com/docker/docker/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type switchCfg struct {
	rpcAddr        string
	traefikGnoFile string
	backupDir      string
	hostPWD        string
	promAddr       string
}

func (c *switchCfg) RegisterFlags(fs *flag.FlagSet) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fs.StringVar(&c.rpcAddr, "rpc-url", "http://rpc.portal.gno.local:26657", "tendermint rpc url")
	fs.StringVar(&c.traefikGnoFile, "traefik-gno-file", "./traefik/gno.yml", "traefik gno file")
	fs.StringVar(&c.backupDir, "backup-dir", "./backups", "backup directory")
	fs.StringVar(&c.hostPWD, "host-pwd", wd, "host pwd (for docker usage)")
	fs.StringVar(&c.promAddr, "prom-addr", ":9090", "listening address for prometheus exporter")
}

func newSwitchCmd(_ commands.IO) *commands.Command {
	cfg := &switchCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "switch",
			ShortUsage: "switch [flags]",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execSwitch(ctx, cfg)
		},
	)
}

func execSwitch(ctx context.Context, cfg *switchCfg) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	portalLoop := &snapshotter{}

	portalLoop, err = NewSnapshotter(dockerClient, config{
		backupDir:      cfg.backupDir,
		rpcAddr:        cfg.rpcAddr,
		hostPWD:        cfg.hostPWD,
		traefikGnoFile: cfg.traefikGnoFile,
	})
	if err != nil {
		return err
	}

	return StartPortalLoop(ctx, portalLoop, true)
}
