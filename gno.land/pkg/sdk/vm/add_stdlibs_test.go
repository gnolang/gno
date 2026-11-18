package vm

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	authm "github.com/gnolang/gno/tm2/pkg/sdk/auth"
	bankm "github.com/gnolang/gno/tm2/pkg/sdk/bank"
	pm "github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addedStdlib stands for a stdlib a release adds: a leaf, which no other
// stdlib imports, so a store can hold every stdlib but it.
const addedStdlib = "encoding/csv"

// chainWithout is a committed store whose genesis loaded every stdlib but
// skip, as a chain started before skip existed.
type chainWithout struct {
	ms               store.CommitMultiStore
	baseKey, iavlKey store.StoreKey
}

func newChainWithout(t *testing.T, skip string) chainWithout {
	t.Helper()

	c := chainWithout{baseKey: store.NewStoreKey("baseCapKey"), iavlKey: store.NewStoreKey("iavlCapKey")}
	db := memdb.NewMemDB()
	c.ms = store.NewCommitMultiStore(db)
	c.ms.MountStoreWithDB(c.baseKey, dbadapter.StoreConstructor, db)
	c.ms.MountStoreWithDB(c.iavlKey, storebptree.FastStoreConstructor, db)
	require.NoError(t, c.ms.LoadLatestVersion())

	vmk, ctx := c.boot(nil)
	acck, bankk, _ := c.keepers()
	acck.SetParams(ctx, authm.DefaultParams())
	bankk.SetParams(ctx, bankm.DefaultParams())
	vmk.SetParams(ctx, DefaultParams())
	ctx = vmk.MakeGnoTransactionStore(ctx)
	gs := vmk.getGnoTransactionStore(ctx)
	for _, lib := range stdlibs.InitOrder() {
		if lib != skip {
			loadStdlibPackage(lib, stdlibsDir, gs)
		}
	}
	vmk.CommitGnoTransactionStore(ctx)
	c.commit(ctx)
	return c
}

func (c chainWithout) keepers() (authm.AccountKeeper, bankm.BankKeeper, pm.ParamsKeeper) {
	prmk := pm.NewParamsKeeper(c.iavlKey)
	acck := authm.NewAccountKeeper(c.iavlKey, prmk.ForModule(authm.ModuleName), std.ProtoBaseAccount, std.ProtoBaseSessionAccount)
	bankk := bankm.NewBankKeeper(acck, prmk.ForModule(bankm.ModuleName), c.iavlKey, []string{"ugnot"})
	prmk.Register(authm.ModuleName, acck)
	prmk.Register(bankm.ModuleName, bankk)
	return acck, bankk, prmk
}

// boot starts a node on the committed store, as NewAppWithOptions does, and
// returns its keeper and a block context over a cache of the store.
func (c chainWithout) boot(migrationStdlibs []string) (*VMKeeper, sdk.Context) {
	acck, bankk, prmk := c.keepers()
	vmk := NewVMKeeper(c.baseKey, c.iavlKey, acck, bankk, prmk)
	prmk.Register(ModuleName, vmk)
	vmk.SetMigrationStdlibs(migrationStdlibs)
	mcw := c.ms.MultiCacheWrap()
	vmk.Initialize(log.NewNoopLogger(), mcw)
	mcw.MultiWrite()
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, c.ms.MultiCacheWrap(), &bft.Header{ChainID: "test-chain-id", Height: 42}, log.NewNoopLogger())
	return vmk, ctx
}

func (c chainWithout) commit(ctx sdk.Context) {
	ctx.MultiStore().(store.MultiStore).MultiWrite()
	c.ms.Commit()
}

// TestAddStdlibs pins a migration adding a stdlib to a running chain: a node
// boots on a store without it only when a migration declares it, the
// migration loads it once, and a realm importing it then costs the same gas
// on the node that ran the migration as on one restarted after it.
func TestAddStdlibs(t *testing.T) {
	t.Parallel()

	c := newChainWithout(t, addedStdlib)

	require.PanicsWithError(t, `intialization error: stdlib "encoding/csv" is not in the store`, func() { c.boot(nil) })

	vmk, ctx := c.boot([]string{addedStdlib})
	require.NoError(t, vmk.AddStdlibs(ctx, stdlibsDir, []string{addedStdlib}))
	require.ErrorContains(t, vmk.AddStdlibs(ctx, stdlibsDir, []string{addedStdlib}), "already in the store")

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	const pkgPath = "gno.land/r/test/csvuser"
	tx := func(vmk *VMKeeper, ctx sdk.Context, msg std.Msg) int64 {
		t.Helper()
		ctx = vmk.MakeGnoTransactionStore(ctx.WithGasMeter(store.NewGasMeter(maxGasQuery)))
		var err error
		switch msg := msg.(type) {
		case MsgAddPackage:
			err = vmk.AddPackage(ctx, msg)
		case MsgCall:
			_, err = vmk.Call(ctx, msg)
		}
		require.NoError(t, err)
		vmk.CommitGnoTransactionStore(ctx)
		return ctx.GasMeter().GasConsumed()
	}
	acck, bankk, _ := c.keepers()
	acck.SetAccount(ctx, acck.NewAccountWithAddress(ctx, addr))
	bankk.SetCoins(ctx, addr, initialBalance)
	tx(vmk, ctx, NewMsgAddPackage(addr, pkgPath, []*std.MemFile{
		{Name: "csvuser.gno", Body: `package csvuser

import (
	"encoding/csv"
	"strings"
)

func Fields(cur realm) int {
	records, err := csv.NewReader(strings.NewReader("a,b,c\nd,e,f\n")).ReadAll()
	if err != nil {
		panic(err)
	}
	return len(records) * len(records[0])
}
`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}))
	// A node's first call after a boot costs more whatever the migration, so
	// each side is measured on its second call.
	call := NewMsgCall(addr, nil, pkgPath, "Fields", nil)
	tx(vmk, ctx, call)
	migrated := tx(vmk, ctx, call)
	c.commit(ctx)

	restarted, ctx := c.boot([]string{addedStdlib})
	tx(restarted, ctx, call)
	assert.Equal(t, migrated, tx(restarted, ctx, call), "gas on the migrated node differs from a restarted one")
}
