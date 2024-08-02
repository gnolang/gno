package doctest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	authm "github.com/gnolang/gno/tm2/pkg/sdk/auth"
	bankm "github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

// Option constants
const (
	IGNORE       = "ignore"       // Do not run the code block
	SHOULD_PANIC = "should_panic" // Expect a panic
	ASSERT       = "assert"       // Assert the result and expected output are equal
)

const (
	goLang  = "go"
	gnoLang = "gno"
)

// GetStdlibsDir returns the path to the standard libraries directory.
func GetStdlibsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot get current file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "stdlibs")
}

// cache stores the results of code execution.
var cache = newCache(maxCacheSize)

// hashCodeBlock generates a SHA256 hash for the given code block.
func hashCodeBlock(c codeBlock) string {
	h := sha256.New()
	h.Write([]byte(c.content))
	return hex.EncodeToString(h.Sum(nil))
}

// ExecuteCodeBlock executes a parsed code block and executes it in a gno VM.
func ExecuteCodeBlock(c codeBlock, stdlibDir string) (string, error) {
	if c.options.Ignore {
		return "IGNORED", nil
	}

	// Extract the actual language from the lang field
	lang := strings.Split(c.lang, ",")[0]

	if lang != goLang && lang != gnoLang {
		return fmt.Sprintf("SKIPPED (Unsupported language: %s)", lang), nil
	}

	if lang == goLang {
		lang = gnoLang
	}

	hashKey := hashCodeBlock(c)

	if result, found := cache.get(hashKey); found {
		result, err := compareResults(result, c.expectedOutput, c.expectedError)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s (cached)", result), nil
	}

	src, err := analyzeAndModifyCode(c.content)
	if err != nil {
		return "", err
	}

	baseKey := store.NewStoreKey("baseKey")
	iavlKey := store.NewStoreKey("iavlKey")

	db := memdb.NewMemDB()

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())
	acck := authm.NewAccountKeeper(iavlKey, std.ProtoBaseAccount)
	bank := bankm.NewBankKeeper(acck)
	stdlibsDir := GetStdlibsDir()
	vmk := vm.NewVMKeeper(baseKey, iavlKey, acck, bank, stdlibsDir, 100_000_000)

	mcw := ms.MultiCacheWrap()
	vmk.Initialize(log.NewNoopLogger(), mcw, true)
	mcw.MultiWrite()

	files := []*std.MemFile{
		{Name: fmt.Sprintf("%d.%s", c.index, lang), Body: src},
	}

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := acck.NewAccountWithAddress(ctx, addr)
	acck.SetAccount(ctx, acc)

	msg2 := vm.NewMsgRun(addr, std.Coins{}, files)

	res, err := vmk.Run(ctx, msg2)
	if c.options.PanicMessage != "" {
		if err == nil {
			return "", fmt.Errorf("expected panic with message: %s, but executed successfully", c.options.PanicMessage)
		}
		if !strings.Contains(err.Error(), c.options.PanicMessage) {
			return "", fmt.Errorf("expected panic with message: %s, but got: %s", c.options.PanicMessage, err.Error())
		}
		return fmt.Sprintf("panicked as expected: %v", err), nil
	}

	if err != nil {
		return "", err
	}

	cache.set(hashKey, res)

	return compareResults(res, c.expectedOutput, c.expectedError)
}

// compareResults compares the actual output of code execution with the expected output or error.
func compareResults(actual, expectedOutput, expectedError string) (string, error) {
	actual = strings.TrimSpace(actual)
	expected := strings.TrimSpace(expectedOutput)
	if expected == "" {
		expected = strings.TrimSpace(expectedError)
	}

	if expected == "" {
		return actual, nil
	}

	if strings.HasPrefix(expected, "regex:") {
		return compareRegex(actual, strings.TrimPrefix(expected, "regex:"))
	}

	if actual != expected {
		return "", fmt.Errorf("expected:\n%s\n\nbut got:\n%s", expected, actual)
	}

	return actual, nil
}

// compareRegex compares the actual output against a regex pattern.
// It returns an error if the regex is invalid or if the actual output does not match the pattern.
func compareRegex(actual, pattern string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	if !re.MatchString(actual) {
		return "", fmt.Errorf("output did not match regex pattern:\npattern: %s\nactual: %s", pattern, actual)
	}

	return actual, nil
}

// ExecuteMatchingCodeBlock executes all code blocks in the given content that match the given pattern.
// It returns a slice of execution results as strings and any error encountered during the execution.
func ExecuteMatchingCodeBlock(ctx context.Context, content string, pattern string) ([]string, error) {
	codeBlocks := GetCodeBlocks(content)
	var results []string

	for _, block := range codeBlocks {
		if matchPattern(block.name, pattern) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				result, err := ExecuteCodeBlock(block, GetStdlibsDir())
				if err != nil {
					return nil, fmt.Errorf("failed to execute code block %s: %w", block.name, err)
				}
				results = append(results, fmt.Sprintf("\n=== %s ===\n\n%s\n", block.name, result))
			}
		}
	}

	return results, nil
}

var (
	regexCache   = make(map[string]*regexp.Regexp)
	regexCacheMu sync.RWMutex
)

// getCompiledRegex retrieves or compiles a regex pattern.
// it uses a cache to store compiled regex patterns for reuse.
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	regexCacheMu.RLock()
	re, exists := regexCache[pattern]
	regexCacheMu.RUnlock()

	if exists {
		return re, nil
	}

	regexCacheMu.Lock()
	defer regexCacheMu.Unlock()

	// double-check in case another goroutine has compiled the regex
	if re, exists = regexCache[pattern]; exists {
		return re, nil
	}

	compiledPattern := regexp.QuoteMeta(pattern)
	compiledPattern = strings.ReplaceAll(compiledPattern, "\\*", ".*")
	re, err := regexp.Compile(compiledPattern)
	if err != nil {
		return nil, err
	}

	regexCache[pattern] = re
	return re, nil
}

// matchPattern checks if a name matches the specific pattern.
func matchPattern(name, pattern string) bool {
	if pattern == "" {
		return true
	}

	re, err := getCompiledRegex(pattern)
	if err != nil {
		return false
	}

	return re.MatchString(name)
}
