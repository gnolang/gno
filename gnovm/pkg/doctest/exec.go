package doctest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

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

// cache stores the results of code execution.
var cache = struct {
	m map[string]string
	sync.RWMutex
}{m: make(map[string]string)}

// hashCodeBlock generates a SHA256 hash for the given code block.
func hashCodeBlock(c CodeBlock) string {
	h := sha256.New()
	h.Write([]byte(c.Content))
	return hex.EncodeToString(h.Sum(nil))
}

// ExecuteCodeBlock executes a parsed code block and executes it in a gno VM.
func ExecuteCodeBlock(c CodeBlock, stdlibDir string) (string, error) {
	if c.T == "go" {
		c.T = "gno"
	} else if c.T != "gno" {
		return "", fmt.Errorf("unsupported language type: %s", c.T)
	}

	hashKey := hashCodeBlock(c)

	// using cached result to avoid re-execution
	cache.RLock()
	result, found := cache.m[hashKey]
	cache.RUnlock()

	if found {
		return fmt.Sprintf("%s (cached)", result), nil
	}

	src, opts, err := analyzeAndModifyCode(c.Content)
	if err != nil {
		return "", err
	}

	if opts.ignore {
		return "[skip]", nil
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

	files := []*std.MemFile{
		{Name: fmt.Sprintf("%d.%s", c.Index, c.T), Body: src},
	}

	coins := std.MustParseCoins("")
	msg2 := vm.NewMsgRun(addr, coins, files)

	res, err := vmk.Run(ctx, msg2)
	if opts.shouldPanic {
		if err == nil {
			return "", fmt.Errorf("expected panic, but code executed successfully")
		}
		return fmt.Sprintf("panicked as expected: %v", err), nil
	}

	if err != nil {
		return "", err
	}

	if opts.expected != "" {
		if !strings.Contains(res, opts.expected) {
			return res, fmt.Errorf("output mismatch: expected to %q, got %q", opts.expected, res)
		}
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
