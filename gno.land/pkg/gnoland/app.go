package gnoland

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

// NewApp creates the GnoLand application.
func NewApp(rootDir string, skipFailingGenesisTxs bool, logger log.Logger) (abci.Application, error) {
	// Get main DB.
	db, err := dbm.NewDB("gnolang", dbm.GoLevelDBBackend, filepath.Join(rootDir, "data"))
	if err != nil {
		return nil, fmt.Errorf("error initializing database %q using path %q: %w", dbm.GoLevelDBBackend, rootDir, err)
	}

	// Capabilities keys.
	mainKey := store.NewStoreKey("main")
	baseKey := store.NewStoreKey("base")

	// Create BaseApp.
	baseApp := sdk.NewBaseApp("gnoland", logger, db, baseKey, mainKey)
	baseApp.SetAppVersion("dev")

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoreWithDB(mainKey, iavl.StoreConstructor, db)
	baseApp.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)

	// Construct keepers.
	acctKpr := auth.NewAccountKeeper(mainKey, ProtoGnoAccount)
	bankKpr := bank.NewBankKeeper(acctKpr)
	stdlibsDir := filepath.Join("..", "gnovm", "stdlibs")
	vmKpr := vm.NewVMKeeper(baseKey, mainKey, acctKpr, bankKpr, stdlibsDir)

	// Set InitChainer
	baseApp.SetInitChainer(InitChainer(baseApp, acctKpr, bankKpr, skipFailingGenesisTxs))

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
				fmt.Println("ERROR LOG:", res.Log)
				fmt.Println("#", i, string(amino.MustMarshalJSON(tx)))
				// NOTE: comment out to ignore.
				if !skipFailingGenesisTxs {
					panic(res.Error)
				}
			} else {
				fmt.Println("SUCCESS:", string(amino.MustMarshalJSON(tx)))
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
