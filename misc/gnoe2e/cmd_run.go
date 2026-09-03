package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/integration"
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
	oracle       bool
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
		case "oracle":
			o.Oracle = &c.oracle
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
		return fmt.Errorf("-mnemonic is required")
	}
	return c.cluster.Validate()
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	// Kept so clusterOverrides can ask which flags were actually given. A
	// scenario declares its own cluster, so a flag left at its default must
	// not override that declaration, and only the FlagSet knows the
	// difference between unset and set to the default value.
	c.flags = fs

	fs.BoolVar(&c.oracle, "oracle", false, "provision the package-approver oracle, overriding what scripts declare")
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
			ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
			defer cancel()

			ctx, cancel = runContext(ctx, cfg.timeout)
			defer cancel()

			return execRun(ctx, cfg, args)
		},
	)
}

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
	// gpaoKey is nil when no scenario asked for an oracle. gpaoAddr is then
	// the zero address.
	gpaoKey  crypto.PrivKey
	gpaoAddr crypto.Address
}

// suite is what every scenario in a run shares: identities and binaries.
type suite struct {
	ids                 runIdentities
	gnolandBin, gpaoBin string
	cleanup             func()
}

func execRun(ctx context.Context, cfg *runCfg, args []string) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	scripts, err := integ.ResolveScriptFiles(args)
	if err != nil {
		return err
	}

	logger := slog.New(termlog.NewHandler(os.Stderr, cfg.verbose))

	scenarios, err := integ.ResolveScenarios(scripts, cfg.clusterOverrides())
	if err != nil {
		return err
	}

	s, err := prepareSuite(ctx, cfg, scenarios, logger)
	if err != nil {
		return err
	}
	defer s.cleanup()

	// Every scenario runs even after one fails. Each owns its cluster, so a
	// failure tells you nothing about the ones after it, and stopping there
	// would hide them.
	var failures []error
	for i, scen := range scenarios {
		logger.Info("scenario", "index", i+1, "of", len(scenarios),
			"script", filepath.Base(scen.Path),
			"validators", scen.Spec.Validators, "oracle", scen.Spec.Oracle)
		rc, teardown, err := prepareScenario(ctx, cfg, s, scen, logger)
		if err == nil {
			err = integ.Run(rc)
			teardown()
		}
		if err != nil {
			logger.Error("scenario failed", "script", filepath.Base(scen.Path), "err", err)
			failures = append(failures, err)
		}
	}
	if err := joinFailures(failures, len(scenarios)); err != nil {
		return err
	}

	logger.Info("all scenarios passed")
	return nil
}

// prepareSuite provisions everything the scenarios share, before any of them
// runs. Its cleanup discards the keybase and the binaries together.
func prepareSuite(ctx context.Context, cfg *runCfg, scenarios []integ.Scenario, logger *slog.Logger) (*suite, error) {
	ids, cleanupIDs, err := setupIdentities(cfg, scenarios, logger)
	if err != nil {
		return nil, err
	}

	// Both binaries are built once for the whole run: a scenario differs from
	// its neighbours in the chain it declares, never in the code that serves
	// it. Their own directory, not the keybase one -- gnokey reads keys out of
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

	var gpaoBin string
	if ids.gpaoKey != nil {
		gpaoBin, err = bldr.Build(ctx, builder.BuildOpts{Binary: "gpao", OutDir: binDir})
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("build gpao: %w", err)
		}
	}

	return &suite{ids: ids, gnolandBin: gnolandBin, gpaoBin: gpaoBin, cleanup: cleanup}, nil
}

// joinFailures reports every scenario that failed rather than only the first,
// so one run answers for the whole suite.
func joinFailures(failures []error, total int) error {
	if len(failures) == 0 {
		return nil
	}
	return fmt.Errorf("%d of %d scenarios failed: %w", len(failures), total, errors.Join(failures...))
}

// setupIdentities imports the run's keys once, since every cluster signs with
// the same accounts.
func setupIdentities(cfg *runCfg, scenarios []integ.Scenario, logger *slog.Logger) (runIdentities, func(), error) {
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

	// Driven by what the scenarios ask for rather than by a flag: deriving a
	// key nobody asked for would fund an account and build a binary for
	// nothing.
	wantsOracle := slices.ContainsFunc(scenarios, func(s integ.Scenario) bool { return s.Spec.Oracle })
	if !wantsOracle {
		return ids, cleanup, nil
	}

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
	ids.gpaoKey = gpaoKey
	ids.gpaoAddr = gpaoKey.PubKey().Address()
	logger.Info("gpao approver provisioned", "address", ids.gpaoAddr.String(), "key", gpaoKeyName)

	return ids, cleanup, nil
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
	if ids.gpaoKey != nil {
		clusterCfg.Genesis.Balances[ids.gpaoAddr.String()] = 1_000_000_000
	}
	if err := scen.Spec.ApplyTo(&clusterCfg, ids.userAddr, ids.gpaoAddr); err != nil {
		return integ.RunConfig{}, nil, err
	}
	if err := clusterCfg.Validate(); err != nil {
		return integ.RunConfig{}, nil, err
	}

	// Said once per scenario, because nothing else about the run will say it:
	// an oracle that is not an approver starts, follows blocks and verifies
	// normally, and every package it approves stays inert. The status board
	// does record the refusal per package, so an operator who already suspects
	// something can find it -- but they have to suspect it first, and the only
	// symptom is deploys that never activate.
	if ids.gpaoKey != nil && !slices.Contains(clusterCfg.Genesis.PkgApprovers, ids.gpaoAddr) {
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

	var gpao integ.GpaoConfig
	if scen.Spec.Oracle && s.gpaoBin != "" {
		gpao = integ.GpaoConfig{
			BinaryPath: s.gpaoBin,
			Mnemonic:   cfg.gpaoMnemonic,
			ChainID:    clusterCfg.Genesis.ChainID,
			Remote:     cl.RPCAddr,
			GnoRoot:    gnoenv.RootDir(),
		}
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
