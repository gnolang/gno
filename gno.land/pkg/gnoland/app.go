package gnoland

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"

	// Only goleveldb is supported for now.
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

type AppOptions struct {
	DB dbm.DB
	// `gnoRootDir` should point to the local location of the gno repository.
	// It serves as the gno equivalent of GOROOT.
	GnoRootDir       string
	GenesisTxHandler GenesisTxHandler
	Logger           *slog.Logger
	MaxCycles        int64
}

func NewAppOptions() *AppOptions {
	return &AppOptions{
		GenesisTxHandler: PanicOnFailingTxHandler,
		Logger:           log.NewNoopLogger(),
		DB:               memdb.NewMemDB(),
		GnoRootDir:       gnoenv.RootDir(),
	}
}

func (c *AppOptions) validate() error {
	if c.Logger == nil {
		return fmt.Errorf("no logger provided")
	}

	if c.DB == nil {
		return fmt.Errorf("no db provided")
	}

	return nil
}

// NewApp creates the GnoLand application.
func NewAppWithOptions(cfg *AppOptions) (abci.Application, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Capabilities keys.
	mainKey := store.NewStoreKey("main")
	baseKey := store.NewStoreKey("base")

	// Create BaseApp.
	// TODO: Add a consensus based min gas prices for the node, by default it does not check
	baseApp := sdk.NewBaseApp("gnoland", cfg.Logger, cfg.DB, baseKey, mainKey)
	baseApp.SetAppVersion("dev")

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoreWithDB(mainKey, iavl.StoreConstructor, cfg.DB)
	baseApp.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, cfg.DB)

	// Construct keepers.
	acctKpr := auth.NewAccountKeeper(mainKey, ProtoGnoAccount)
	bankKpr := bank.NewBankKeeper(acctKpr)

	// XXX: Embed this ?
	stdlibsDir := filepath.Join(cfg.GnoRootDir, "gnovm", "stdlibs")
	vmKpr := vm.NewVMKeeper(baseKey, mainKey, acctKpr, bankKpr, stdlibsDir, cfg.MaxCycles)

	// Set InitChainer
	baseApp.SetInitChainer(InitChainer(baseApp, acctKpr, bankKpr, cfg.GenesisTxHandler))

	// Set AnteHandler
	authOptions := auth.AnteOptions{
		VerifyGenesisSignatures: false, // for development
	}
	authAnteHandler := auth.NewAnteHandler(
		acctKpr, bankKpr, auth.DefaultSigVerificationGasConsumer, authOptions)
	baseApp.SetAnteHandler(
		// Override default AnteHandler with custom logic.
		func(ctx sdk.Context, tx std.Tx, simulate bool) (
			newCtx sdk.Context, res sdk.Result, abort bool,
		) {
			// Override auth params.
			ctx = ctx.WithValue(
				auth.AuthParamsContextKey{}, auth.DefaultParams())
			// Continue on with default auth ante handler.
			newCtx, res, abort = authAnteHandler(ctx, tx, simulate)
			return
		},
	)

	// Set EndBlocker
	baseApp.SetEndBlocker(EndBlocker(vmKpr))

	// Set a handler Route.
	baseApp.Router().AddRoute("auth", auth.NewHandler(acctKpr))
	baseApp.Router().AddRoute("bank", bank.NewHandler(bankKpr))
	baseApp.Router().AddRoute("vm", vm.NewHandler(vmKpr))

	// Load latest version.
	if err := baseApp.LoadLatestVersion(); err != nil {
		return nil, err
	}

	// Initialize the VMKeeper.
	vmKpr.Initialize(baseApp.GetCacheMultiStore())

	return baseApp, nil
}

// NewApp creates the GnoLand application.
func NewApp(dataRootDir string, skipFailingGenesisTxs bool, logger *slog.Logger, maxCycles int64) (abci.Application, error) {
	var err error

	cfg := NewAppOptions()
	if skipFailingGenesisTxs {
		cfg.GenesisTxHandler = NoopGenesisTxHandler
	}

	// Get main DB.
	cfg.DB, err = dbm.NewDB("gnolang", dbm.GoLevelDBBackend, filepath.Join(dataRootDir, config.DefaultDBDir))
	if err != nil {
		return nil, fmt.Errorf("error initializing database %q using path %q: %w", dbm.GoLevelDBBackend, dataRootDir, err)
	}

	cfg.Logger = logger
	return NewAppWithOptions(cfg)
}

type GenesisTxHandler func(ctx sdk.Context, tx std.Tx, res sdk.Result)

func NoopGenesisTxHandler(ctx sdk.Context, tx std.Tx, res sdk.Result) {}

func PanicOnFailingTxHandler(ctx sdk.Context, tx std.Tx, res sdk.Result) {
	if res.IsErr() {
		panic(res.Log)
	}
}

// InitChainer returns a function that can initialize the chain with genesis.
func InitChainer(baseApp *sdk.BaseApp, acctKpr auth.AccountKeeperI, bankKpr bank.BankKeeperI, resHandler GenesisTxHandler) func(sdk.Context, abci.RequestInitChain) abci.ResponseInitChain {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		// Get genesis state.
		genState := req.AppState.(GnoGenesisState)
		// Parse and set genesis state balances.
		for _, bal := range genState.Balances {
			acc := acctKpr.NewAccountWithAddress(ctx, bal.Address)
			acctKpr.SetAccount(ctx, acc)
			err := bankKpr.SetCoins(ctx, bal.Address, bal.Amount)
			if err != nil {
				panic(err)
			}
		}

		// Run genesis txs.
		for _, tx := range genState.Txs {
			res := baseApp.Deliver(tx)
			if res.IsErr() {
				ctx.Logger().Error(
					"Unable to deliver genesis tx",
					"log", res.Log,
					"error", res.Error,
					"gas-used", res.GasUsed,
				)
			}

			resHandler(ctx, tx, res)
		}

		// Done!
		return abci.ResponseInitChain{
			Validators: req.Validators,
		}
	}
}

// XXX not used yet.
func EndBlocker(vmk vm.VMKeeperI) func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		return abci.ResponseEndBlock{}
	}
}
