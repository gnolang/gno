package auth

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type testEnv struct {
	ctx   sdk.Context
	acck  AccountKeeper
	bankk BankKeeperI
	gk    GasPriceKeeper
}

func setupTestEnv() testEnv {
	db := memdb.NewMemDB()

	authCapKey := store.NewStoreKey("authCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()
	prmk := params.NewParamsKeeper(authCapKey)

	acck := NewAccountKeeper(authCapKey, prmk.ForModule(ModuleName), std.ProtoBaseAccount)
	bankk := NewDummyBankKeeper(acck, prmk.ForModule("dummybank"))
	gk := NewGasPriceKeeper(authCapKey)

	prmk.Register(ModuleName, acck)
	prmk.Register("dummybank", bankk)

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{Height: 1, ChainID: "test-chain-id"}, log.NewNoopLogger())

	acck.SetParams(ctx, DefaultParams()) // Setup default params

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

	return testEnv{ctx: ctx, acck: acck, bankk: bankk, gk: gk}
}

// DummyBankKeeper defines a supply keeper used only for testing to avoid
// circle dependencies
type DummyBankKeeper struct {
	acck AccountKeeper
}

// NewDummyBankKeeper creates a DummyBankKeeper instance
func NewDummyBankKeeper(acck AccountKeeper, prmk params.ParamsKeeperI) DummyBankKeeper {
	return DummyBankKeeper{acck}
}

func (bankk DummyBankKeeper) SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	return bankk.SendCoins(ctx, fromAddr, toAddr, amt)
}

// SendCoins for the dummy supply keeper
func (bankk DummyBankKeeper) SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	fromAcc := bankk.acck.GetAccount(ctx, fromAddr)
	toAcc := bankk.acck.GetAccount(ctx, toAddr)
	if toAcc == nil {
		toAcc = bankk.acck.NewAccountWithAddress(ctx, toAddr)
	}

	newFromCoins := fromAcc.GetCoins().SubUnsafe(amt)
	if !newFromCoins.IsValid() {
		return std.ErrInsufficientCoins(fromAcc.GetCoins().String())
	}
	newToCoins := toAcc.GetCoins().Add(amt)
	if err := fromAcc.SetCoins(newFromCoins); err != nil {
		return std.ErrInternal(err.Error())
	}
	bankk.acck.SetAccount(ctx, fromAcc)
	if err := toAcc.SetCoins(newToCoins); err != nil {
		return std.ErrInternal(err.Error())
	}
	bankk.acck.SetAccount(ctx, toAcc)

	return nil
}

// WillSetParam checks if the key contains the module's parameter key prefix and updates the module parameter accordingly.
func (bankk DummyBankKeeper) WillSetParam(ctx sdk.Context, key string, value any) {}
