package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/misc/gnoe2e/internal/builder"
	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	integ "github.com/gnolang/gno/misc/gnoe2e/internal/integration"
	"github.com/gnolang/gno/misc/gnoe2e/internal/termlog"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

const (
	defaultKeyName  = "test1"
	defaultMnemonic = integration.DefaultAccount_Seed

	// gpaoKeyName is the keybase name the oracle's derived key is imported
	// under, so scenarios can reference it (e.g. gnokey) without deriving it
	// themselves.
	gpaoKeyName = "gpao"
	// defaultGpaoMnemonic is the standard BIP39 test vector. The oracle derives
	// its signer at account 0 index 0 with no way to change that, so it needs a
	// mnemonic of its own rather than an index in the run's.
	defaultGpaoMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
)

type runCfg struct {
	keyName      string
	mnemonic     string
	gpaoMnemonic string
	timeout      time.Duration
	verbose      bool
	cluster      cluster.ClusterConfig

	// flags is the set the command was parsed with, so clusterOverrides can
	// tell a flag that was given from one left at its default.
	flags *flag.FlagSet
}

// clusterOverrides collects the cluster settings named on the command line.
//
// flag.FlagSet.Visit reports only flags that were set, which is the whole
// point: a scenario declares its own cluster, and "-validators 2" has to beat
// that declaration while a -validators left alone must not.
//
// -pkg-approver is repeatable and a spec holds one approver, so only the last
// is honoured here. Nothing passes it more than once.
func (c *runCfg) clusterOverrides() integ.ClusterOverrides {
	var o integ.ClusterOverrides
	if c.flags == nil {
		return o
	}
	c.flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "validators":
			o.Validators = &c.cluster.NumValidators
		case "code-submission-policy":
			policy := string(c.cluster.Genesis.CodeSubmissionPolicy)
			o.CodeSubmissionPolicy = &policy
		case "block-max-gas":
			o.BlockMaxGas = &c.cluster.Genesis.MaxGas
		case "pkg-approver":
			if approvers := c.cluster.Genesis.PkgApprovers; len(approvers) > 0 {
				addr := approvers[len(approvers)-1].String()
				o.PkgApprover = &addr
			}
		}
	})
	return o
}

func (c *runCfg) validate() error {
	if c.mnemonic == "" {
		return errors.New("-mnemonic is required")
	}
	return c.cluster.Validate()
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	// Kept so clusterOverrides can ask which flags were actually given. A
	// scenario declares its own cluster, so a flag left at its default must
	// not override that declaration, and only the FlagSet knows the
	// difference between unset and set to the default value.
	c.flags = fs

	fs.StringVar(&c.keyName, "keyname", c.keyName, "key name for test account")
	fs.StringVar(&c.mnemonic, "mnemonic", c.mnemonic, "mnemonic for key derivation")
	fs.StringVar(&c.gpaoMnemonic, "gpao-mnemonic", c.gpaoMnemonic, "mnemonic for the package-approver oracle key")
	fs.DurationVar(&c.timeout, "timeout", c.timeout, "maximum duration of the whole run; 0 for no limit")
	fs.BoolVar(&c.verbose, "verbose", false, "verbose output")
	c.cluster.RegisterFlags(fs)
}

// defaultRunCfg is the run a bare "gnoe2e run" performs, and the source of the
// flag defaults. Separate from newRunCmd so a go test driver, which registers
// no flags, starts from the same settings the command line does.
func defaultRunCfg() *runCfg {
	clusterCfg := cluster.DefaultClusterConfig()
	clusterCfg.Genesis.LoadExamples = false // run deploys packages via testscript commands, not genesis

	return &runCfg{
		keyName:      defaultKeyName,
		mnemonic:     defaultMnemonic,
		gpaoMnemonic: defaultGpaoMnemonic,
		timeout:      10 * time.Minute,
		cluster:      clusterCfg,
	}
}

func newRunCmd(_ commands.IO) *commands.Command {
	cfg := defaultRunCfg()

	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] [scripts-dir]",
			ShortHelp:  "run txtar test scripts against a gnoland cluster",
			LongHelp:   "Runs txtar test scripts against a local validator cluster",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			ctx, cancel := signal.NotifyContext(ctx, runSignals...)
			defer cancel()

			ctx, cancel = runContext(ctx, cfg.timeout)
			defer cancel()

			return execRun(ctx, cfg, args)
		},
	)
}

// newRunLogger builds the logger a run reports through and installs it as the
// slog default.
//
// The default is where internal/cluster's own diagnostics go: they are
// package-level slog calls, and left to the handler slog starts with they are
// dropped below Info and formatted like nothing else in the run.
func newRunLogger(w io.Writer, verbose bool) *slog.Logger {
	logger := slog.New(termlog.NewHandler(w, verbose))
	slog.SetDefault(logger)
	return logger
}

// runSignals end a run: the interrupt a terminal sends, and the SIGTERM a CI
// runner sends when it cancels a job or a user sends with a bare kill. Both
// have to reach the run's own teardown, because the validators and the temp
// directories are its to remove and nothing else will.
var runSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// runContext bounds the whole run rather than each scenario: a suite that
// hangs has to end by itself, and a scenario booting its own cluster has no
// deadline of its own to hit.
//
// The node processes are started from this context, so reaching the deadline
// kills them and the scenario they were serving fails. It does not reach the
// script itself: one parked in sleep, or waiting inside a gnokey call, runs to
// its own end and fails on the cluster that went away under it.
//
// A timeout of zero means no limit, as it does for `go test -timeout`.
func runContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

// deriveGpaoKey derives the oracle's signing key from the run's oracle
// mnemonic.
//
// One derivation, used both for the genesis approver entry and for the keybase
// import, so the chain can never end up with an approver nobody can sign for.
func deriveGpaoKey(mnemonic string) (crypto.PrivKey, error) {
	key, err := integration.GeneratePrivKeyFromMnemonic(mnemonic, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("derive gpao key: %w", err)
	}
	return key, nil
}

// runIdentities are the addresses and keys every cluster in a run shares.
type runIdentities struct {
	gnoHome  string
	keyName  string
	userAddr crypto.Address
	// gpaoAddr is the oracle's approver address, provisioned for every run
	// whether or not a script starts the oracle.
	gpaoAddr crypto.Address
}

// suite is what every scenario in a run shares: identities and binaries.
type suite struct {
	ids        runIdentities
	gnolandBin string
	// gpaoBin yields the oracle binary. Deferred rather than built with
	// gnoland, because whether the oracle runs at all is a decision the
	// scripts make with "gpao start".
	gpaoBin func() (string, error)
	cleanup func()
}

func execRun(ctx context.Context, cfg *runCfg, args []string) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	scripts, err := integ.ResolveScriptFiles(args)
	if err != nil {
		return err
	}

	logger := newRunLogger(os.Stderr, cfg.verbose)

	scenarios, err := integ.ResolveScenarios(scripts, cfg.clusterOverrides())
	if err != nil {
		return err
	}

	s, err := prepareSuite(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer s.cleanup()

	err = runScenarios(ctx, scenarios, logger, func(scen integ.Scenario) error {
		rc, teardown, err := prepareScenario(ctx, cfg, s, scen, logger)
		if err != nil {
			return err
		}
		defer teardown()
		return integ.Run(rc)
	})
	if err != nil {
		return err
	}

	logger.Info("all scenarios passed")
	return nil
}

// runScenarios drives each scenario in turn and reports how many failed.
//
// Every scenario runs even after one fails. Each owns its cluster, so a failure
// tells you nothing about the ones after it, and stopping there would hide
// them. A cancelled run is the exception: past its deadline, or past a Ctrl-C,
// every remaining scenario would boot a temp dir, keys and a genesis before
// os/exec refused to start a process under a dead context, and would be
// reported as a failure it never had the chance to have.
func runScenarios(
	ctx context.Context,
	scenarios []integ.Scenario,
	logger *slog.Logger,
	run func(integ.Scenario) error,
) error {
	var failures []error
	attempted := 0
	for i, scen := range scenarios {
		if err := ctx.Err(); err != nil {
			// Reported as a failure of the run whether or not the scenarios it
			// got through passed: a suite cut off partway has not answered the
			// question it was asked, and exiting 0 says it has.
			logger.Warn("run cancelled; the remaining scenarios were not attempted",
				"remaining", len(scenarios)-i, "err", err)
			return fmt.Errorf("run cancelled after %d of %d scenarios, %d of them failed: %w",
				attempted, len(scenarios), len(failures), err)
		}
		attempted++
		logger.Info("scenario", "index", i+1, "of", len(scenarios),
			"script", filepath.Base(scen.Path),
			"validators", scen.Spec.Validators)
		if err := run(scen); err != nil {
			logger.Error("scenario failed", "script", filepath.Base(scen.Path), "err", err)
			failures = append(failures, err)
		}
	}
	return failureSummary(failures, attempted)
}

// prepareSuite provisions everything the scenarios share, before any of them
// runs. Its cleanup discards the keybase and the binaries together.
func prepareSuite(ctx context.Context, cfg *runCfg, logger *slog.Logger) (*suite, error) {
	ids, cleanupIDs, err := setupIdentities(cfg, logger)
	if err != nil {
		return nil, err
	}

	// A binary is built once for the whole run: a scenario differs from its
	// neighbours in the chain it declares, never in the code that serves it.
	// Their own directory, not the keybase one -- gnokey reads keys out of
	// that.
	binDir, err := os.MkdirTemp("", "gnoe2e-bin-*")
	if err != nil {
		cleanupIDs()
		return nil, fmt.Errorf("create binary dir: %w", err)
	}
	cleanup := func() {
		os.RemoveAll(binDir)
		cleanupIDs()
	}

	bldr := builder.NewLocalBuilder()
	gnolandBin, err := bldr.Build(ctx, builder.BuildOpts{OutDir: binDir})
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("build gnoland: %w", err)
	}

	gpaoBin := sync.OnceValues(func() (string, error) {
		path, err := bldr.Build(ctx, builder.BuildOpts{Binary: "gpao", OutDir: binDir})
		if err != nil {
			return "", fmt.Errorf("build gpao: %w", err)
		}
		return path, nil
	})

	return &suite{ids: ids, gnolandBin: gnolandBin, gpaoBin: gpaoBin, cleanup: cleanup}, nil
}

// failureSummary reports how much of the suite failed.
//
// Counted against what was attempted rather than what was listed, because a run
// cut short by its deadline stops partway and the scenarios it never reached
// failed nothing. The failures themselves are not repeated here: each was
// reported where it happened, with the script that caused it, and a failure
// printed twice reads as two.
func failureSummary(failures []error, attempted int) error {
	if len(failures) == 0 {
		return nil
	}
	return fmt.Errorf("%d of %d scenarios failed", len(failures), attempted)
}

// setupIdentities imports the run's keys once, since every cluster signs with
// the same accounts.
func setupIdentities(cfg *runCfg, logger *slog.Logger) (runIdentities, func(), error) {
	gnoHomeDir, err := os.MkdirTemp("", "gnoe2e-home-*")
	if err != nil {
		return runIdentities{}, nil, fmt.Errorf("failed to create gnohome dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(gnoHomeDir) }

	privKey, err := integration.GeneratePrivKeyFromMnemonic(cfg.mnemonic, "", 0, 0)
	if err != nil {
		cleanup()
		return runIdentities{}, nil, fmt.Errorf("failed to derive key from mnemonic: %w", err)
	}

	kb, err := keys.NewKeyBaseFromDir(gnoHomeDir)
	if err != nil {
		cleanup()
		return runIdentities{}, nil, fmt.Errorf("failed to create keybase: %w", err)
	}
	if err := kb.ImportPrivKey(cfg.keyName, privKey, ""); err != nil {
		cleanup()
		return runIdentities{}, nil, fmt.Errorf("failed to import key: %w", err)
	}

	ids := runIdentities{
		gnoHome:  gnoHomeDir,
		keyName:  cfg.keyName,
		userAddr: privKey.PubKey().Address(),
	}
	logger.Info("key imported", "name", cfg.keyName, "address", ids.userAddr.String())

	gpaoKey, err := deriveGpaoKey(cfg.gpaoMnemonic)
	if err != nil {
		cleanup()
		return runIdentities{}, nil, err
	}
	// The approver also needs to be a usable gnokey identity: the control arm
	// of an oracle scenario activates a package directly, which separates a
	// package the chain refuses from one only the oracle refused.
	if err := kb.ImportPrivKey(gpaoKeyName, gpaoKey, ""); err != nil {
		cleanup()
		return runIdentities{}, nil, fmt.Errorf("import gpao key: %w", err)
	}
	ids.gpaoAddr = gpaoKey.PubKey().Address()
	logger.Info("gpao approver provisioned", "address", ids.gpaoAddr.String(), "key", gpaoKeyName)

	return ids, cleanup, nil
}

// oracleCannotActivate reports the one misconfiguration that leaves the oracle
// looking healthy: an inert chain whose approver set does not hold its key.
// The oracle starts, follows blocks and verifies normally, and every package it
// approves stays inert. A chain that is not inert parks nothing, so its
// approver set says nothing about the oracle.
func oracleCannotActivate(genesis cluster.GenesisConfig, oracle crypto.Address) bool {
	// Read from the resolved params rather than the named fields: the genesis.
	// family is applied after them, so a scenario can set both the policy and
	// the approver set through a path without either field changing.
	//
	// A genesis that will not resolve is not this function's to report. The
	// same resolution runs again inside StartCluster, where the key that broke
	// it is named.
	state, err := cluster.ResolveGenesisState(genesis)
	if err != nil {
		return false
	}
	return state.VM.Params.CodeSubmissionPolicy == vm.CodeSubmissionPolicyInert &&
		!slices.Contains(state.VM.Params.PkgApprovers, oracle)
}

// prepareScenario boots the cluster one scenario declared and assembles the
// run that will drive its script. The returned func takes the cluster down
// again, and the caller owes it that call once the script is done.
//
// Built from genesis for this scenario and discarded after, so nothing it
// deploys, spends or commits can reach another one.
func prepareScenario(
	ctx context.Context,
	cfg *runCfg,
	s *suite,
	scen integ.Scenario,
	logger *slog.Logger,
) (integ.RunConfig, func(), error) {
	ids := s.ids
	clusterCfg := cfg.cluster
	clusterCfg.Genesis.Balances = maps.Clone(cfg.cluster.Genesis.Balances)
	clusterCfg.Genesis.Balances[ids.userAddr.String()] = 1_000_000_000 // 1000 GNOT
	clusterCfg.Genesis.Balances[ids.gpaoAddr.String()] = 1_000_000_000
	if err := scen.Spec.ApplyTo(&clusterCfg, ids.userAddr, ids.gpaoAddr); err != nil {
		return integ.RunConfig{}, nil, err
	}
	if err := clusterCfg.Validate(); err != nil {
		return integ.RunConfig{}, nil, err
	}

	// Said once per scenario, because nothing else about the run will say it.
	// The status board does record the refusal per package, so an operator who
	// already suspects something can find it -- but they have to suspect it
	// first, and the only symptom is deploys that never activate.
	if oracleCannotActivate(clusterCfg.Genesis, ids.gpaoAddr) {
		logger.Warn("oracle key is not a package approver; it will verify packages but never activate them",
			"oracle", ids.gpaoAddr.String())
	}

	clusterCfg.Logger = logger.With("component", "cluster")
	cl, err := cluster.StartCluster(ctx, clusterCfg, s.gnolandBin)
	if err != nil {
		return integ.RunConfig{}, nil, fmt.Errorf("start cluster: %w", err)
	}

	rpcAddrs := make([]string, len(cl.Validators))
	for i, v := range cl.Validators {
		rpcAddrs[i] = v.RPCAddr
	}
	logger.Info("cluster ready", "rpc_addr", cl.RPCAddr, "validators", len(cl.Validators))

	gpao := integ.GpaoConfig{
		Binary:   s.gpaoBin,
		Mnemonic: cfg.gpaoMnemonic,
		ChainID:  clusterCfg.Genesis.ChainID,
		Remote:   cl.RPCAddr,
		GnoRoot:  gnoenv.RootDir(),
	}

	return integ.RunConfig{
		ScriptPath:  scen.Path,
		RPCAddr:     cl.RPCAddr,
		RPCAddrs:    rpcAddrs,
		ChainID:     clusterCfg.Genesis.ChainID,
		GnoHome:     ids.gnoHome,
		UserAddr:    ids.userAddr.String(),
		KeyName:     ids.keyName,
		Verbose:     cfg.verbose,
		Logger:      logger,
		Gpao:        gpao,
		GpaoKeyName: gpaoKeyName,
		GpaoAddr:    ids.gpaoAddr.String(),
		Cluster:     cl,
	}, cl.Cleanup, nil
}
