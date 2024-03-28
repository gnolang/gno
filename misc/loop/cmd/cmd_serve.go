package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type serveCfg struct {
	rpcAddr        string
	traefikGnoFile string
	backupDir      string
	hostPWD        string
	promAddr       string
}

type serveService struct {
	cfg serveCfg

	// TODO(albttx): put getter on it with RMutex
	portalLoop *snapshotter

	portalLoopURL string
}

func (c *serveCfg) RegisterFlags(fs *flag.FlagSet) {
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

func newServeCmd(commands.IO) *commands.Command {
	cfg := &serveCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "serve",
			ShortUsage: "serve [flags]",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execServe(ctx, cfg, args)
		},
	)
}

func execServe(ctx context.Context, cfg *serveCfg, _ []string) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	portalLoop := &snapshotter{}

	// Serve monitoring
	go func() {
		s := &monitoringService{
			portalLoop: portalLoop,
		}

		for portalLoop.url == "" {
			time.Sleep(time.Second * 1)
		}

		go s.recordMetrics()

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.promAddr, nil)
	}()

	// the loop
	for {
		portalLoop, err = NewSnapshotter(dockerClient, config{
			backupDir:      cfg.backupDir,
			rpcAddr:        cfg.rpcAddr,
			hostPWD:        cfg.hostPWD,
			traefikGnoFile: cfg.traefikGnoFile,
		})
		if err != nil {
			return err
		}

		err = StartPortalLoop(ctx, portalLoop, false)
		if err != nil {
			logrus.WithError(err).Error()
		}
		time.Sleep(time.Second * 120)
	}
}
