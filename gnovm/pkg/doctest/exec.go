package doctest

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	bankm "github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// Option constants
const (
	IGNORE       = "ignore"       // Do not run the code block
	SHOULD_PANIC = "should_panic" // Expect a panic
	ASSERT       = "assert"       // Assert the result and expected output are equal
)

const STDLIBS_DIR = "../../stdlibs"

// ExecuteCodeBlock executes a parsed code block and executes it in a gno VM.
func ExecuteCodeBlock(c CodeBlock, stdlibDir string) (string, error) {
	if c.T == "go" {
		c.T = "gno"
	} else if c.T != "gno" {
		return "", fmt.Errorf("unsupported language type: %s", c.T)
	}

	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")

	ms, ctx := setupMultiStore(baseCapKey, iavlCapKey)

	acck := auth.NewAccountKeeper(iavlCapKey, std.ProtoBaseAccount)
	bank := bankm.NewBankKeeper(acck)

	vmk := vm.NewVMKeeper(baseCapKey, iavlCapKey, acck, bank, stdlibDir, 100_000_000)
	vmk.Initialize(ms.MultiCacheWrap())

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := acck.NewAccountWithAddress(ctx, addr)
	acck.SetAccount(ctx, acc)

	files := []*std.MemFile{
		{Name: fmt.Sprintf("%d.%s", c.Index, c.T), Body: c.Content},
	}

	coins := std.MustParseCoins("")
	msg2 := vm.NewMsgRun(addr, coins, files)

	res, err := vmk.Run(ctx, msg2)
	if err != nil {
		return "", err
	}

	return res, nil
}

func setupMultiStore(baseKey, iavlKey types.StoreKey) (types.CommitMultiStore, sdk.Context) {
	db := memdb.NewMemDB()

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "chain-id"}, log.NewNoopLogger())
	return ms, ctx
}
