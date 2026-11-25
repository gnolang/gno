package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gnolang/gno/contribs/gnosupervisor/internal/oci"
	"github.com/gnolang/gno/contribs/gnosupervisor/internal/supervisor"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type runCfg struct {
	ledger      string
	dataDir     string
	genesis     string
	cacheDir    string
	platform    string
	registry    string
	stopTimeout time.Duration
	maxRestarts int
	logFormat   string
	continueAt  int64
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.ledger, "ledger", "", "path to the chain's upgrades.json (required)")
	fs.StringVar(&c.dataDir, "data-dir", "gnoland-data", "gnoland data directory, passed to gnoland start")
	fs.StringVar(&c.genesis, "genesis", "genesis.json", "genesis file the ledger records, passed to gnoland start")
	fs.StringVar(&c.cacheDir, "cache-dir", "gnosupervisor-cache", "where release images are extracted")
	fs.StringVar(&c.platform, "platform", runtime.GOOS+"/"+runtime.GOARCH, "image platform to pull, os/arch")
	fs.StringVar(&c.registry, "registry", "", "registry base URL replacing https://<host> of the ledger's image references (a mirror)")
	fs.DurationVar(&c.stopTimeout, "stop-timeout", 60*time.Second, "how long gnoland gets to exit after SIGTERM")
	fs.IntVar(&c.maxRestarts, "max-restarts", 5, "unexpected exits in a row, with no block committed in between, before giving up")
	fs.StringVar(&c.logFormat, "log-format", "text", "supervisor log format: text or json")
	fs.Int64Var(&c.continueAt, "continue-after-halt", 0, "go on, on the same release, after a stop at this height that the ledger does not record as a halt (a pause)")
}

var platformRE = regexp.MustCompile(`^[a-z0-9]+/[a-z0-9]+$`)

// Flags the supervisor sets on gnoland start itself.
var ownedFlags = []string{"data-dir", "genesis", "gnoroot-dir"}

func newRunCmd(io commands.IO) *commands.Command {
	cfg := &runCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run --ledger <upgrades.json> [flags] [-- <gnoland start flags>]",
			ShortHelp:  "run a node, switching release at each halt height the ledger records",
			LongHelp: `Runs gnoland start as a child process, on the release the ledger assigns to the
node's next block. When the child stops at its range's halt height, the next
release starts on the same data directory. Releases are pulled from their
image digests in the ledger and verified; no container runtime is used.

Arguments after -- are passed to gnoland start, except --data-dir, --genesis
and --gnoroot-dir, which the supervisor sets.`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execRun(ctx, cfg, args, io)
		},
	)
}

func execRun(ctx context.Context, c *runCfg, extra []string, io commands.IO) error {
	if c.ledger == "" {
		return errors.New("--ledger is required")
	}
	if !platformRE.MatchString(c.platform) {
		return fmt.Errorf("--platform %q is not os/arch", c.platform)
	}
	for _, a := range extra {
		name := strings.TrimLeft(strings.SplitN(a, "=", 2)[0], "-")
		for _, owned := range ownedFlags {
			if strings.HasPrefix(a, "-") && name == owned {
				return fmt.Errorf("--%s is set by the supervisor; use the supervisor's own flag", owned)
			}
		}
	}

	var handler slog.Handler
	switch c.logFormat {
	case "text":
		handler = slog.NewTextHandler(io.Err(), nil)
	case "json":
		handler = slog.NewJSONHandler(io.Err(), nil)
	default:
		return fmt.Errorf("--log-format %q: want text or json", c.logFormat)
	}
	log := slog.New(handler).With("component", "gnosupervisor")

	dataDir, err := filepath.Abs(c.dataDir)
	if err != nil {
		return err
	}
	genesis, err := filepath.Abs(c.genesis)
	if err != nil {
		return err
	}
	cacheDir, err := filepath.Abs(c.cacheDir)
	if err != nil {
		return err
	}

	plan, err := supervisor.LoadPlan(c.ledger, genesis)
	if err != nil {
		return err
	}
	lock, err := supervisor.Lock(dataDir)
	if err != nil {
		return err
	}
	defer lock.Close()
	node, err := supervisor.OpenNode(dataDir)
	if err != nil {
		return err
	}
	if err := node.CheckChain(plan.ChainID()); err != nil {
		return err
	}

	client := oci.NewClient()
	client.BaseURL = c.registry
	log.Info("supervising", "chain_id", plan.ChainID(), "data_dir", dataDir, "cache_dir", cacheDir)
	live := node.LiveHeight()
	if live == nil {
		log.Warn("the node's RPC does not listen on TCP: a halt reached outside block sync will not be noticed")
	}

	return supervisor.Run(ctx, supervisor.Config{
		Plan: plan,
		Node: node,
		Artifacts: &supervisor.Store{
			Dir:      cacheDir,
			Platform: c.platform,
			Client:   client,
			Log:      log,
		},
		DataDir:     dataDir,
		Genesis:     genesis,
		ExtraArgs:   extra,
		LockFile:    lock,
		StopTimeout: c.stopTimeout,
		MaxRestarts: c.maxRestarts,
		BackoffMin:  2 * time.Second,
		BackoffMax:  time.Minute,

		Live:              live,
		HaltParam:         node.HaltParam(),
		PollInterval:      5 * time.Second,
		StuckAfter:        2 * time.Minute,
		StallAfter:        10 * time.Minute,
		ContinueAfterHalt: c.continueAt,

		Stdout: io.Out(),
		Stderr: io.Err(),
		Log:    log,
	})
}
