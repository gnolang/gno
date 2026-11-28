package bank

// DONTCOVER

import (
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type testEnv struct {
	ctx   sdk.Context
	bankk BankKeeper
	acck  auth.AccountKeeper
	key   store.StoreKey
	prmk  params.ParamsKeeper
}

// testAccountDenom stands in for the chain's gas denom. tm2 itself has no
// native-denom concept — gno.land supplies the allowlist — so tests name it
// rather than assuming it.
const testAccountDenom = "ugnot"

// accountTierTestDenoms mirrors what gno.land passes in: only the gas denom is
// held in the account object. Every other denom in these tests is split-tier.
var accountTierTestDenoms = []string{testAccountDenom}

func setupTestEnv() testEnv {
	db := memdb.NewMemDB()

	authCapKey := store.NewStoreKey("authCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())

	prmk := params.NewParamsKeeper(authCapKey)
	acck := auth.NewAccountKeeper(authCapKey, prmk.ForModule(auth.ModuleName), std.ProtoBaseAccount, std.ProtoBaseSessionAccount)
	bankk := NewBankKeeper(acck, prmk.ForModule(ModuleName), authCapKey, accountTierTestDenoms)

	prmk.Register(auth.ModuleName, acck)
	prmk.Register(ModuleName, bankk)

	return testEnv{ctx: ctx, bankk: bankk, acck: acck, prmk: prmk, key: authCapKey}
}
