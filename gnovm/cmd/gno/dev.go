package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type devCfg struct {
	verbose  bool
	bindAddr string
}

var defaultDevOptions = &devCfg{
	verbose: false,
}

func newDevCmd(io commands.IO) *commands.Command {
	cfg := &devCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "dev",
			ShortUsage: "dev [flags]",
			ShortHelp:  "Devs run a node for dev purpose",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDev(cfg, args, io)
		},
	)
}

func (c *devCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"verbose",
		defaultDevOptions.verbose,
		"verbose output when deving",
	)

	fs.StringVar(
		&c.bindAddr,
		"web-bind",
		"127.0.0.1:8888",
		"verbose output when deving",
	)
}

func execDev(cfg *devCfg, args []string, io commands.IO) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup trap signal
	osm.TrapSignal(cancel)

	logger := log.NewNopLogger()
	// logger := log.NewTMLogger(log.NewSyncWriter(io.Out))

	rootdir := gnoland.MustGuessGnoRootDir()

	// GNOWeb setup
	var server http.Server
	{
		webConfig := gnoweb.NewDefaultConfig()
		app := gnoweb.MakeApp(webConfig)

		server.ReadHeaderTimeout = 60 * time.Second
		server.Handler = app.Router

		l, err := net.Listen("tcp", cfg.bindAddr)
		if err != nil {
			return fmt.Errorf("unable to listen to %q: %w", cfg.bindAddr, err)
		}

		go func() {
			if err := server.Serve(l); err != nil {
				io.ErrPrintfln("HTTP server stopped with error: %v", err)
			}
		}()

		io.Printfln("gnoweb listening on %q", l.Addr().String())
		defer server.Close()
	}

	// Setup node
	var node *node.Node
	{
		var err error

		nodeConfig := gnoland.NewDefaultInMemoryNodeConfig(rootdir)
		nodeConfig.Genesis.AppState = gnoland.GnoGenesisState{
			Balances: []gnoland.Balance{
				{
					Address: crypto.MustAddressFromString(integration.DefaultAccount_Address), // test1
					Amount:  std.MustParseCoins("10000000000000ugnot"),
				},
			},
			Txs: []std.Tx{},
		}

		node, err = gnoland.NewInMemoryNode(logger, nodeConfig)
		if err != nil {
			return fmt.Errorf("unable to init node: %w", err)
		}

		// Start our node
		if err := node.Start(); err != nil {
			return fmt.Errorf("unable to start node: %w", err)
		}

		defer node.Stop()
	}

	// Wait for readiness
	select {
	case <-ctx.Done():
		return fmt.Errorf("unable to wait for node readiness: %w", ctx.Err())
	case <-gnoland.WaitForNodeReadiness(node): // ok
	}

	io.Println("dev node is now ready")

	// Wait until exit
	<-ctx.Done()

	return nil
}
