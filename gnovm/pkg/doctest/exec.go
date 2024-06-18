package doctest

import (
	"fmt"

	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
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

const (
	IGNORE       = "ignore"
	SHOULD_PANIC = "should_panic"
	NO_RUN       = "no_run"
)

const LIBS_DIR = "../../stdlibs"

func ExecuteCodeBlock(c CodeBlock) (string, error) {
	if c.ContainsOptions(IGNORE) {
		return "", nil
	}

	err := validateCodeBlock(c)
	if err != nil {
		return "", err
	}

	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")

	ms, ctx := setupMultiStore(baseCapKey, iavlCapKey)

	acck := auth.NewAccountKeeper(iavlCapKey, std.ProtoBaseAccount)
	bank := bankm.NewBankKeeper(acck)

	vmk := vmm.NewVMKeeper(baseCapKey, iavlCapKey, acck, bank, LIBS_DIR, 100_000_000)

	vmk.Initialize(ms.MultiCacheWrap())

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := acck.NewAccountWithAddress(ctx, addr)
	acck.SetAccount(ctx, acc)

	files := []*std.MemFile{
		{Name: fmt.Sprintf("%d.%s", c.Index, c.T), Body: c.Content},
	}

	coins := std.MustParseCoins("")
	msg2 := vmm.NewMsgRun(addr, coins, files)
	res, err := vmk.Run(ctx, msg2)
	if err != nil {
		return "", err
	}

	return res, nil
}

func validateCodeBlock(c CodeBlock) error {
	if c.T == "go" {
		c.T = "gno"
	} else if c.T != "gno" {
		return fmt.Errorf("unsupported language: %s", c.T)
	}
	return nil
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
