package gnoland

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type AppOptions struct {
	DB dbm.DB
	// `gnoRootDir` should point to the local location of the gno repository.
	// It serves as the gno equivalent of GOROOT.
	GnoRootDir            string
	SkipFailingGenesisTxs bool
	Logger                log.Logger
	MaxCycles             int64
}

func NewAppOptions() *AppOptions {
	return &AppOptions{
		Logger:     log.NewNopLogger(),
		DB:         dbm.NewMemDB(),
		GnoRootDir: GuessGnoRootDir(),
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
	baseApp := sdk.NewBaseApp("gnoland", cfg.Logger, cfg.DB, baseKey, mainKey)
	baseApp.SetAppVersion("dev")

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoreWithDB(mainKey, iavl.StoreConstructor, cfg.DB)
	baseApp.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, cfg.DB)

	// Construct keepers.
	acctKpr := auth.NewAccountKeeper(mainKey, ProtoGnoAccount)
	bankKpr := bank.NewBankKeeper(acctKpr)
	stdlibsDir := filepath.Join(cfg.GnoRootDir, "gnovm", "stdlibs")
	vmKpr := vm.NewVMKeeper(baseKey, mainKey, acctKpr, bankKpr, stdlibsDir, cfg.MaxCycles)

	// Set InitChainer
	baseApp.SetInitChainer(InitChainer(baseApp, acctKpr, bankKpr, cfg.SkipFailingGenesisTxs))

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
func NewApp(dataRootDir string, skipFailingGenesisTxs bool, logger log.Logger, maxCycles int64) (abci.Application, error) {
	var err error

	cfg := NewAppOptions()

	// Get main DB.
	cfg.DB, err = dbm.NewDB("gnolang", dbm.GoLevelDBBackend, filepath.Join(dataRootDir, "data"))
	if err != nil {
		return nil, fmt.Errorf("error initializing database %q using path %q: %w", dbm.GoLevelDBBackend, dataRootDir, err)
	}

	cfg.Logger = logger

	return NewAppWithOptions(cfg)
}

// InitChainer returns a function that can initialize the chain with genesis.
func InitChainer(baseApp *sdk.BaseApp, acctKpr auth.AccountKeeperI, bankKpr bank.BankKeeperI, skipFailingGenesisTxs bool) func(sdk.Context, abci.RequestInitChain) abci.ResponseInitChain {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		// Get genesis state.
		genState := req.AppState.(GnoGenesisState)
		// Parse and set genesis state balances.
		for _, bal := range genState.Balances {
			addr, coins := parseBalance(bal)
			acc := acctKpr.NewAccountWithAddress(ctx, addr)
			acctKpr.SetAccount(ctx, acc)
			err := bankKpr.SetCoins(ctx, addr, coins)
			if err != nil {
				panic(err)
			}
		}
		// Run genesis txs.
		for i, tx := range genState.Txs {
			res := baseApp.Deliver(tx)
			if res.IsErr() {
				ctx.Logger().Error("LOG", res.Log)
				ctx.Logger().Error("#", i, string(amino.MustMarshalJSON(tx)))

				// NOTE: comment out to ignore.
				if !skipFailingGenesisTxs {
					panic(res.Error)
				}
			} else {
				ctx.Logger().Info("SUCCESS:", string(amino.MustMarshalJSON(tx)))
			}
		}
		// Done!
		return abci.ResponseInitChain{
			Validators: req.Validators,
		}
	}
}

func parseBalance(bal string) (crypto.Address, std.Coins) {
	parts := strings.Split(bal, "=")
	if len(parts) != 2 {
		panic(fmt.Sprintf("invalid balance string %s", bal))
	}
	addr, err := crypto.AddressFromBech32(parts[0])
	if err != nil {
		panic(fmt.Sprintf("invalid balance addr %s (%v)", bal, err))
	}
	coins, err := std.ParseCoins(parts[1])
	if err != nil {
		panic(fmt.Sprintf("invalid balance coins %s (%v)", bal, err))
	}
	return addr, coins
}

// XXX not used yet.
func EndBlocker(vmk vm.VMKeeperI) func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		return abci.ResponseEndBlock{}
	}
}

func GuessGnoRootDir() string {
	var rootdir string

	// First try to get the root directory from the GNOROOT environment variable.
	if rootdir = os.Getenv("GNOROOT"); rootdir != "" {
		return filepath.Clean(rootdir)
	}

	if gobin, err := exec.LookPath("go"); err == nil {
		// If GNOROOT is not set, try to guess the root directory using the `go list` command.
		cmd := exec.Command(gobin, "list", "-m", "-mod=mod", "-f", "{{.Dir}}", "github.com/gnolang/gno")
		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Errorf("invalid gno directory %q: %w", rootdir, err))
		}

		return strings.TrimSpace(string(out))
	}

	panic("no go binary available, unable to determine gno root-dir path")
}
