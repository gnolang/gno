package gnoland

import (
	"fmt"
	"path/filepath"
	"strings"

	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/crypto"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/log"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/sdk/auth"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store"
	"github.com/gnolang/gno/pkgs/store/iavl"
)

// NewApp creates the GnoLand application.
func NewApp(rootDir string, logger log.Logger) (abci.Application, error) {
	// Get main DB.
	db := dbm.NewDB("gnolang", dbm.GoLevelDBBackend, filepath.Join(rootDir, "data"))

	// Capabilities key to access the main Store.
	mainKey := store.NewStoreKey(sdk.MainStoreKey)

	// Create BaseApp.
	baseApp := sdk.NewBaseApp("gnoland", logger, db)

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoreWithDB(mainKey, iavl.StoreConstructor, db)

	// Construct keepers.
	authKpr := auth.NewAccountKeeper(mainKey, ProtoGnoAccount)
	bankKpr := bank.NewBankKeeper(authKpr)
	vmKpr := vm.NewVMKeeper(mainKey, authKpr, bankKpr)

	// Configure InitChainer for genesis.
	baseApp.SetInitChainer(InitChainer(authKpr, bankKpr))

	// Set a handler Route.
	baseApp.Router().AddRoute("bank", bank.NewHandler(bankKpr))
	baseApp.Router().AddRoute("vm", vm.NewHandler(vmKpr))

	// Load latest version.
	if err := baseApp.LoadLatestVersion(mainKey); err != nil {
		return nil, err
	}

	return baseApp, nil
}

// InitChainer returns a function that can initialize the chain with genesis.
func InitChainer(authKpr auth.AccountKeeperI, bankKpr bank.BankKeeperI) func(sdk.Context, abci.RequestInitChain) abci.ResponseInitChain {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		// Get genesis state.
		genState := req.AppState.(GnoGenesisState)
		// Parse and set genesis state balances.
		for _, bal := range genState.Balances {
			addr, coins := parseBalance(bal)
			err := bankKpr.SetCoins(ctx, addr, coins)
			if err != nil {
				panic(err)
			}
		}
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
