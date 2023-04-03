package auth

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type testEnv struct {
	ctx  sdk.Context
	acck AccountKeeper
	bank BankKeeperI
}

// moduleAccount defines an account for modules that holds coins on a pool
type moduleAccount struct {
	*std.BaseAccount
	name        string   `json:"name" yaml:"name"`              // name of the module
	permissions []string `json:"permissions" yaml"permissions"` // permissions of module account
}

// HasPermission returns whether or not the module account has permission.
func (ma moduleAccount) HasPermission(permission string) bool {
	for _, perm := range ma.permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetName returns the the name of the holder's module
func (ma moduleAccount) GetName() string {
	return ma.name
}

// GetPermissions returns permissions granted to the module account
func (ma moduleAccount) GetPermissions() []string {
	return ma.permissions
}

func setupTestEnv() testEnv {
	db := dbm.NewMemDB()

	authCapKey := store.NewStoreKey("authCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	acck := NewAccountKeeper(authCapKey, std.ProtoBaseAccount)
	bank := NewDummyBankKeeper(acck)

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{Height: 1, ChainID: "test-chain-id"}, log.NewNopLogger())
	ctx = ctx.WithValue(AuthParamsContextKey{}, DefaultParams())
	ctx = ctx.WithConsensusParams(&abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes:    1024,
			MaxDataBytes:  1024 * 100,
			MaxBlockBytes: 1024 * 100,
			MaxGas:        10 * 1000 * 1000,
			TimeIotaMS:    10,
		},
		Validator: &abci.ValidatorParams{
			PubKeyTypeURLs: []string{}, // XXX
		},
	})

	return testEnv{ctx: ctx, acck: acck, bank: bank}
}

// DummyBankKeeper defines a supply keeper used only for testing to avoid
// circle dependencies
type DummyBankKeeper struct {
	acck AccountKeeper
}

// NewDummyBankKeeper creates a DummyBankKeeper instance
func NewDummyBankKeeper(acck AccountKeeper) DummyBankKeeper {
	return DummyBankKeeper{acck}
}

// SendCoins for the dummy supply keeper
func (bank DummyBankKeeper) SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	fromAcc := bank.acck.GetAccount(ctx, fromAddr)
	toAcc := bank.acck.GetAccount(ctx, toAddr)
	if toAcc == nil {
		toAcc = bank.acck.NewAccountWithAddress(ctx, toAddr)
	}

	newFromCoins := fromAcc.GetCoins().SubUnsafe(amt)
	if !newFromCoins.IsValid() {
		return std.ErrInsufficientCoins(fromAcc.GetCoins().String())
	}
	newToCoins := toAcc.GetCoins().Add(amt)
	if err := fromAcc.SetCoins(newFromCoins); err != nil {
		return std.ErrInternal(err.Error())
	}
	bank.acck.SetAccount(ctx, fromAcc)
	if err := toAcc.SetCoins(newToCoins); err != nil {
		return std.ErrInternal(err.Error())
	}
	bank.acck.SetAccount(ctx, toAcc)

	return nil
}
