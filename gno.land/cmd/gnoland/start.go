package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/events"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/telemetry"

	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultNodeDir = "gnoland-data"

var errMissingGenesis = errors.New("missing genesis.json")

var startGraphic = strings.ReplaceAll(`
                    __             __
  ___ ____  ___    / /__ ____  ___/ /
 / _ '/ _ \/ _ \_ / / _ '/ _ \/ _  /
 \_, /_//_/\___(_)_/\_,_/_//_/\_,_/
/___/
`, "'", "`")

// Keep in sync with contribs/gnogenesis/internal/txs/txs_add_packages.go
var genesisDeployFee = std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1)))

type nodeCfg struct {
	gnoRootDir                 string // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	skipFailingGenesisTxs      bool   // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	skipGenesisSigVerification bool   // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	genesisBalancesFile        string // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	genesisTxsFile             string // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	genesisRemote              string // TODO: remove as part of https://github.com/gnolang/gno/issues/1952
	genesisFile                string
	chainID                    string
	dataDir                    string
	lazyInit                   bool

	logLevel  string
	logFormat string
}

func newStartCmd(io commands.IO) *commands.Command {
	cfg := &nodeCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "start",
			ShortUsage: "start [flags]",
			ShortHelp:  "starts the Gnoland blockchain node",
			LongHelp:   "Starts the Gnoland blockchain node, with accompanying setup",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execStart(ctx, cfg, io)
		},
	)
}

func (c *nodeCfg) RegisterFlags(fs *flag.FlagSet) {
	gnoroot := gnoenv.RootDir()
	defaultGenesisBalancesFile := filepath.Join(gnoroot, "gno.land", "genesis", "genesis_balances.txt")

	fs.BoolVar(
		&c.skipFailingGenesisTxs,
		"skip-failing-genesis-txs",
		false,
		"don't panic when replaying invalid genesis txs",
	)

	fs.BoolVar(
		&c.skipGenesisSigVerification,
		"skip-genesis-sig-verification",
		false,
		"don't panic when replaying invalidly signed genesis txs",
	)

	fs.StringVar(
		&c.genesisBalancesFile,
		"genesis-balances-file",
		defaultGenesisBalancesFile,
		"initial distribution file",
	)

	fs.StringVar(
		&c.genesisTxsFile,
		"genesis-txs-file",
		"",
		"initial txs to replay",
	)

	fs.StringVar(
		&c.genesisFile,
		"genesis",
		"genesis.json",
		"the path to the genesis.json",
	)

	fs.StringVar(
		&c.chainID,
		"chainid",
		"dev",
		"the ID of the chain",
	)

	fs.StringVar(
		&c.gnoRootDir,
		"gnoroot-dir",
		gnoroot,
		"the root directory of the gno repository",
	)

	fs.StringVar(
		&c.dataDir,
		"data-dir",
		defaultNodeDir,
		"the path to the node's data directory",
	)

	fs.StringVar(
		&c.genesisRemote,
		"genesis-remote",
		"localhost:26657",
		"replacement for '%%REMOTE%%' in genesis",
	)

	fs.StringVar(
		&c.logLevel,
		"log-level",
		zapcore.DebugLevel.String(),
		"log level for the gnoland node,",
	)

	fs.StringVar(
		&c.logFormat,
		"log-format",
		log.ConsoleFormat.String(),
		"log format for the gnoland node",
	)

	fs.BoolVar(
		&c.lazyInit,
		"lazy",
		false,
		"flag indicating if lazy init is enabled. Generates the node secrets, configuration, and genesis.json",
	)
}

func execStart(ctx context.Context, c *nodeCfg, io commands.IO) error {
	gnoNode, err := createNode(c, io)
	if err != nil {
		return err
	}

	// Start the node (async)
	if err := gnoNode.Start(); err != nil {
		return fmt.Errorf("unable to start the Gnoland node, %w", err)
	}

	// Wait for the exit signal
	<-ctx.Done()

	if !gnoNode.IsRunning() {
		return nil
	}

	// Gracefully stop the gno node
	if err := gnoNode.Stop(); err != nil {
		return fmt.Errorf("unable to gracefully stop the Gnoland node, %w", err)
	}

	// Gracefully stop the app
	if err = cfg.LocalApp.Close(); err != nil {
		return fmt.Errorf("unable to gracefully close the Gnoland application: %w", err)
	}

	return nil
}

func createNode(c *nodeCfg, io commands.IO) (*node.Node, error) {
	// Get the absolute path to the node's data directory
	nodeDir, err := filepath.Abs(c.dataDir)
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path for data directory, %w", err)
	}

	// Get the absolute path to the node's genesis.json
	genesisPath, err := filepath.Abs(c.genesisFile)
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path for the genesis.json, %w", err)
	}

	// Initialize the logger
	zapLogger, err := initializeLogger(io.Out(), c.logLevel, c.logFormat)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize zap logger, %w", err)
	}

	defer func() {
		// Sync the logger before exiting
		_ = zapLogger.Sync()
	}()

	// Wrap the zap logger
	logger := log.ZapLoggerToSlog(zapLogger)

	if c.lazyInit {
		if err := lazyInitNodeDir(io, nodeDir); err != nil {
			return nil, fmt.Errorf("unable to lazy-init the node directory, %w", err)
		}
	}

	// Load the configuration
	cfg, err := config.LoadConfig(nodeDir)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", tryConfigInit, err)
	}

	// Check if the genesis.json exists
	if !osm.FileExists(genesisPath) {
		if !c.lazyInit {
			return nil, errMissingGenesis
		}

		// Load the private validator secrets
		privateKey := privval.LoadFilePV(
			cfg.PrivValidatorKeyFile(),
			cfg.PrivValidatorStateFile(),
		)

		// Init a new genesis.json
		if err := lazyInitGenesis(io, c, genesisPath, privateKey.Key.PrivKey); err != nil {
			return nil, fmt.Errorf("unable to initialize genesis.json, %w", err)
		}
	}

	// Initialize telemetry
	if err := telemetry.Init(*cfg.Telemetry); err != nil {
		return nil, fmt.Errorf("unable to initialize telemetry, %w", err)
	}

	// Print the starting graphic
	if c.logFormat != string(log.JSONFormat) {
		io.Println(startGraphic)
	}

	// Create a top-level shared event switch
	evsw := events.NewEventSwitch()
	minGasPrices := cfg.Application.MinGasPrices

	// Create application and node
	cfg.LocalApp, err = gnoland.NewApp(
		nodeDir,
		gnoland.GenesisAppConfig{
			SkipFailingTxs:      c.skipFailingGenesisTxs,
			SkipSigVerification: c.skipGenesisSigVerification,
		},
		evsw,
		logger,
		minGasPrices,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create the Gnoland app, %w", err)
	}

	// Create a default node, with the given setup
	gnoNode, err := node.DefaultNewNode(cfg, genesisPath, evsw, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create the Gnoland node, %w", err)
	}

	return gnoNode, err
}

// lazyInitNodeDir initializes new secrets, and a default configuration
// in the given node directory, if not present
func lazyInitNodeDir(io commands.IO, nodeDir string) error {
	var (
		configPath  = constructConfigPath(nodeDir)
		secretsPath = constructSecretsPath(nodeDir)
	)

	// Check if the configuration already exists
	if !osm.FileExists(configPath) {
		// Create the gnoland config options
		cfg := &configInitCfg{
			configCfg: configCfg{
				configPath: constructConfigPath(nodeDir),
			},
		}

		// Run gnoland config init
		if err := execConfigInit(cfg, io); err != nil {
			return fmt.Errorf("unable to initialize config, %w", err)
		}

		io.Printfln("WARN: Initialized default node config at %q", filepath.Dir(cfg.configPath))
		io.Println()
	}

	// Create the gnoland secrets options
	secrets := &secretsInitCfg{
		commonAllCfg: commonAllCfg{
			dataDir: secretsPath,
		},
		forceOverwrite: false, // existing secrets shouldn't be pruned
	}

	// Run gnoland secrets init
	err := execSecretsInit(secrets, []string{}, io)
	if err == nil {
		io.Printfln("WARN: Initialized default node secrets at %q", secrets.dataDir)

		return nil
	}

	// Check if the error is valid
	if errors.Is(err, errOverwriteNotEnabled) {
		// No new secrets were generated
		return nil
	}

	return fmt.Errorf("unable to initialize secrets, %w", err)
}

// lazyInitGenesis a new genesis.json file, with a single validator
func lazyInitGenesis(
	io commands.IO,
	c *nodeCfg,
	genesisPath string,
	signer gnoland.GenesisSigner,
) error {
	// Check if the genesis.json is present
	if osm.FileExists(genesisPath) {
		return nil
	}

	// Generate the new genesis.json file
	if err := generateGenesisFile(genesisPath, signer, c); err != nil {
		return fmt.Errorf("unable to generate genesis file, %w", err)
	}

	io.Printfln("WARN: Initialized genesis.json at %q", genesisPath)

	return nil
}

// initializeLogger initializes the zap logger using the given format and log level,
// outputting to the given IO
func initializeLogger(io io.WriteCloser, logLevel, logFormat string) (*zap.Logger, error) {
	// Initialize the log level
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("unable to parse log level, %w", err)
	}

	// Initialize the log format
	format := log.Format(strings.ToLower(logFormat))

	// Initialize the zap logger
	return log.GetZapLoggerFn(format)(io, level), nil
}

func generateGenesisFile(genesisFile string, privKey crypto.PrivKey, c *nodeCfg) error {
	var (
		pubKey = signer.PubKey()
		// There is an active constraint for gno.land transactions:
		//
		// All transaction messages' (MsgSend, MsgAddPkg...) "author" field,
		// specific to the message type ("creator", "sender"...), must match
		// the signature address contained in the transaction itself.
		// This means that if MsgSend is originating from address A,
		// the owner of the private key for address A needs to sign the transaction
		// containing the message. Every message in a transaction needs to
		// originate from the same account that signed the transaction
		txSender = pubKey.Address()
	)

	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Now()
	gen.ChainID = c.chainID
	gen.ConsensusParams = abci.ConsensusParams{
		Block: &abci.BlockParams{
			// TODO: update limits.
			MaxTxBytes:   1_000_000,     // 1MB,
			MaxDataBytes: 2_000_000,     // 2MB,
			MaxGas:       3_000_000_000, // 3B gas
			TimeIotaMS:   100,           // 100ms
		},
	}

	gen.Validators = []bft.GenesisValidator{
		{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   10,
			Name:    "testvalidator",
		},
	}

	// Load balances files
	balances, err := gnoland.LoadGenesisBalancesFile(c.genesisBalancesFile)
	if err != nil {
		return fmt.Errorf("unable to load genesis balances file %q: %w", c.genesisBalancesFile, err)
	}

	// Load examples folder
	examplesDir := filepath.Join(c.gnoRootDir, "examples")
	pkgsTxs, err := gnoland.LoadPackagesFromDir(examplesDir, txSender, genesisDeployFee)
	if err != nil {
		return fmt.Errorf("unable to load examples folder: %w", err)
	}

	// Load Genesis TXs
	var genesisTxs []gnoland.TxWithMetadata

	if c.genesisTxsFile != "" {
		genesisTxs, err = gnoland.LoadGenesisTxsFile(c.genesisTxsFile, c.chainID, c.genesisRemote)
		if err != nil {
			return fmt.Errorf("unable to load genesis txs file: %w", err)
		}
	}

	genesisTxs = append(pkgsTxs, genesisTxs...)

	// Sign genesis transactions, with the default key (test1)
	if err = gnoland.SignGenesisTxs(genesisTxs, signer, c.chainID); err != nil {
		return fmt.Errorf("unable to sign genesis txs: %w", err)
	}

	// Make sure the genesis transaction author has sufficient
	// balance to cover transaction deployments in genesis.
	//
	// During the init-chainer process, the account that authors the
	// genesis transactions needs to have a sufficient balance
	// to cover outstanding transaction costs.
	// Since the cost can't be estimated upfront at this point, the balance
	// set is an arbitrary value based on a "best guess" basis.
	// There should be a larger discussion if genesis transactions should consume gas, at all
	deployerBalance := int64(len(genesisTxs)) * 2_100_000 // ~2.1 GNOT per tx
	balances.Set(txSender, std.NewCoins(std.NewCoin("ugnot", deployerBalance)))

	// Construct genesis AppState.
	defaultGenState := gnoland.DefaultGenState()
	defaultGenState.Balances = balances.List()
	defaultGenState.Txs = genesisTxs
	gen.AppState = defaultGenState

	// Write genesis state
	if err := gen.SaveAs(genesisFile); err != nil {
		return fmt.Errorf("unable to write genesis file %q: %w", genesisFile, err)
	}

	return nil
}
