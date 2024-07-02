package doctest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
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

// cache stores the results of code execution.
var cache struct {
	m map[string]string
	sync.RWMutex
}

func init() {
	cache.m = make(map[string]string)
}

// hashCodeBlock generates a SHA256 hash for the given code block.
func hashCodeBlock(c codeBlock) string {
	h := sha256.New()
	h.Write([]byte(c.content))
	return hex.EncodeToString(h.Sum(nil))
}

// ExecuteCodeBlock executes a parsed code block and executes it in a gno VM.
func ExecuteCodeBlock(c codeBlock, stdlibDir string) (string, error) {
	if c.lang == "go" {
		c.lang = "gno"
	} else if c.lang != "gno" {
		return "", fmt.Errorf("unsupported language type: %s", c.lang)
	}

	hashKey := hashCodeBlock(c)

	// using cached result to avoid re-execution
	cache.RLock()
	result, found := cache.m[hashKey]
	cache.RUnlock()

	if found {
		return fmt.Sprintf("%s (cached)", result), nil
	}

	src, err := analyzeAndModifyCode(c.content)
	if err != nil {
		return "", err
	}

	baseKey := store.NewStoreKey("baseKey")
	iavlKey := store.NewStoreKey("iavlKey")

	ms, ctx := setupMultiStore(baseKey, iavlKey)

	acck := auth.NewAccountKeeper(iavlKey, std.ProtoBaseAccount)
	bank := bankm.NewBankKeeper(acck)

	vmk := vm.NewVMKeeper(baseKey, iavlKey, acck, bank, stdlibDir, 100_000_000)
	vmk.Initialize(ms.MultiCacheWrap())

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := acck.NewAccountWithAddress(ctx, addr)
	acck.SetAccount(ctx, acc)

	memPkg := &std.MemPackage{
		Name:  "main",
		Path:  "main",
		Files: []*std.MemFile{{Name: fmt.Sprintf("%d.%s", c.index, c.lang), Body: src}},
	}

	getter := newDynPackageLoader(stdlibDir)
	if err := gnolang.TypeCheckMemPackage(memPkg, getter); err != nil {
		return "", fmt.Errorf("type checking failed: %w", err)
	}

	files := []*std.MemFile{
		{Name: fmt.Sprintf("%d.%s", c.index, c.lang), Body: src},
	}

	coins := std.MustParseCoins("")
	msg2 := vm.NewMsgRun(addr, coins, files)

	res, err := vmk.Run(ctx, msg2)
	if err != nil {
		return "", err
	}

	cache.Lock()
	cache.m[hashKey] = res
	cache.Unlock()

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
