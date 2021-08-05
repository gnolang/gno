package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/classic/abci/types"
	"github.com/tendermint/classic/crypto/secp256k1"
	dbm "github.com/tendermint/classic/db"
	"github.com/tendermint/classic/libs/log"
	tmtypes "github.com/tendermint/classic/types"

	"github.com/tendermint/classic/sdk/store"
	"github.com/tendermint/classic/sdk/x/auth"
	"github.com/tendermint/classic/sdk/x/bank"
	"github.com/tendermint/classic/sdk/x/params"
	"github.com/tendermint/classic/sdk/x/supply/internal/types"

	sdk "github.com/tendermint/classic/sdk/types"
)

// nolint: deadcode unused
var (
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
	holder     = "holder"
)

// nolint: deadcode unused
func createTestInput(t *testing.T, isCheckTx bool, initPower int64, nAccs int64) (sdk.Context, auth.AccountKeeper, Keeper) {

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keySupply := sdk.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "supply-chain"}, isCheckTx, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)

	blacklistedAddrs := make(map[string]bool)

	pk := params.NewKeeper(keyParams, tkeyParams, params.DefaultCodespace)
	ak := auth.NewAccountKeeper(keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)

	valTokens := sdk.TokensFromConsensusPower(initPower)

	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens))
	createTestAccs(ctx, int(nAccs), initialCoins, &ak)

	maccPerms := map[string][]string{
		holder:       nil,
		types.Minter: []string{types.Minter},
		types.Burner: []string{types.Burner},
		multiPerm:    []string{types.Minter, types.Burner, types.Staking},
		randomPerm:   []string{"random"},
	}
	keeper := NewKeeper(keySupply, ak, bk, maccPerms)
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens.MulRaw(nAccs)))
	keeper.SetSupply(ctx, types.NewSupply(totalSupply))

	return ctx, ak, keeper
}

// nolint: unparam deadcode unused
func createTestAccs(ctx sdk.Context, numAccs int, initialCoins sdk.Coins, ak *auth.AccountKeeper) (accs []auth.Account) {
	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())
		acc := auth.NewBaseAccountWithAddress(addr)
		acc.Coins = initialCoins
		acc.PubKey = pubKey
		acc.AccountNumber = uint64(i)
		ak.SetAccount(ctx, &acc)
	}
	return
}
