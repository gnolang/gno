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

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	authm "github.com/gnolang/gno/tm2/pkg/sdk/auth"
	bankm "github.com/gnolang/gno/tm2/pkg/sdk/bank"
	paramsm "github.com/gnolang/gno/tm2/pkg/sdk/params"
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
	gnoLang      = "gno"
)

var (
	cache      = newCache(maxCacheSize)
	regexCache = make(map[string]*regexp.Regexp)

	addrRegex = regexp.MustCompile(`gno\.land/[pre]/[a-z0-9]+/[a-z_/.]+`)
)

// ExecuteCodeBlock executes a parsed code block and executes it in a gno VM.
func ExecuteCodeBlock(c codeBlock, stdlibDir string) (string, error) {
	if c.options.Ignore {
		return "IGNORED", nil
	}

	// Extract the actual language from the lang field
	lang := strings.Split(c.lang, ",")[0]
	if lang != gnoLang {
		return fmt.Sprintf("SKIPPED (Unsupported language: %s)", lang), nil
	}

	hashKey := hashCodeBlock(c)

	// get the result from the cache if it exists
	if result, found := cache.get(hashKey); found {
		return handleCachedResult(result, c)
	}

	ctx, acck, _, vmk, stdlibCtx := setupEnv()

	files := []*std.MemFile{
		{Name: fmt.Sprintf("%d.%s", c.index, lang), Body: c.content},
	}

	// create a freash account for the code block
	privKey := ed25519.GenPrivKey()
	addr := privKey.PubKey().Address()
	acc := acck.NewAccountWithAddress(ctx, addr)
	acck.SetAccount(ctx, acc)

	msg2 := vm.NewMsgRun(addr, std.Coins{}, files)

	res, err := vmk.Run(stdlibCtx, msg2)
	if c.options.PanicMessage != "" {
		return handlePanicMessage(err, c.options.PanicMessage)
	}

	// remove package path from the result and replace with `main`.
	res = replacePackagePath(res)

	if err != nil {
		return "", err
	}

	cache.set(hashKey, res)

	// If there is no expected output or error, It is considered
	// a simple code execution and the result is returned as is.
	if c.expectedOutput == "" && c.expectedError == "" {
		return res, nil
	}

	// Otherwise, compare the actual output with the expected output or error.
	return compareResults(res, c.expectedOutput, c.expectedError)
}

// ExecuteMatchingCodeBlock executes all code blocks in the given content that match the given pattern.
// It returns a slice of execution results as strings and any error encountered during the execution.
func ExecuteMatchingCodeBlock(
	ctx context.Context,
	content string,
	pattern string,
) ([]string, error) {
	codeBlocks, err := GetCodeBlocks(content)
	if err != nil {
		return nil, err
	}

	results := make([]string, 0, len(codeBlocks))
	for _, block := range codeBlocks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if !matchPattern(block.name, pattern) {
			continue
		}

		result, err := ExecuteCodeBlock(block, GetStdlibsDir())
		if err != nil {
			return nil, fmt.Errorf("failed to execute code block %s: %w", block.name, err)
		}
		results = append(results, fmt.Sprintf("\n=== %s ===\n\n%s\n", block.name, result))
	}

	return results, nil
}

// setupEnv creates and initializes the execution environment for running extracted code blocks.
// It sets up necessary keepers (account, bank, VM), initializes a test chain context,
// and loads standard libraries. The function returns the context, keepers, and stdlib context
// needed for code execution.
//
// ref: gno.land/pkg/sdk/vm/common_test.go
func setupEnv() (
	sdk.Context,
	authm.AccountKeeper,
	bankm.BankKeeper,
	*vm.VMKeeper,
	sdk.Context,
) {
	baseKey := store.NewStoreKey("baseKey")
	iavlKey := store.NewStoreKey("iavlKey")

	db := memdb.NewMemDB()

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(
		sdk.RunTxModeDeliver,
		ms,
		&bft.Header{ChainID: "test-chain-id"},
		log.NewNoopLogger(),
	)
	prmk := paramsm.NewParamsKeeper(iavlKey)
	acck := authm.NewAccountKeeper(iavlKey, prmk.ForModule(authm.ModuleName), std.ProtoBaseAccount)
	bank := bankm.NewBankKeeper(acck, prmk.ForModule(bankm.ModuleName))

	prmk.Register(authm.ModuleName, acck)
	prmk.Register(bankm.ModuleName, bank)

	mcw := ms.MultiCacheWrap()

	vmk := vm.NewVMKeeper(baseKey, iavlKey, acck, bank, prmk)
	prmk.Register(vm.ModuleName, vmk)
	vmk.SetParams(ctx, vm.DefaultParams())
	vmk.Initialize(log.NewNoopLogger(), mcw)

	stdlibCtx := vmk.MakeGnoTransactionStore(ctx.WithMultiStore(mcw))
	stdlibsDir := GetStdlibsDir()
	vmk.LoadStdlib(stdlibCtx, stdlibsDir)
	vmk.CommitGnoTransactionStore(stdlibCtx)

	mcw.MultiWrite()

	return ctx, acck, bank, vmk, stdlibCtx
}

func handleCachedResult(result string, c codeBlock) (string, error) {
	res := strings.TrimSpace(result)

	if c.expectedOutput == "" && c.expectedError == "" {
		return fmt.Sprintf("%s (cached)", res), nil
	}

	res, err := compareResults(res, c.expectedOutput, c.expectedError)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s (cached)", res), nil
}

func handlePanicMessage(err error, panicMessage string) (string, error) {
	if err == nil {
		return "", fmt.Errorf(
			"expected panic with message: %s, but executed successfully",
			panicMessage,
		)
	}

	if strings.Contains(err.Error(), panicMessage) {
		return fmt.Sprintf("panicked as expected: %v", err), nil
	}

	return "", fmt.Errorf(
		"expected panic with message: %s, but got: %s",
		panicMessage, err.Error(),
	)
}

// compareResults compares the actual output of code execution with the expected output or error.
func compareResults(actual, expectedOutput, expectedError string) (string, error) {
	actual = strings.TrimSpace(actual)
	expected := strings.TrimSpace(expectedOutput)
	if expected == "" {
		expected = strings.TrimSpace(expectedError)
	}

	if expected == "" {
		if actual != "" {
			return "", fmt.Errorf("expected no output, but got:\n%s", actual)
		}
		return "", nil
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
		return "", fmt.Errorf(
			"output did not match regex pattern:\npattern: %s\nactual: %s",
			pattern, actual,
		)
	}

	return actual, nil
}

// getCompiledRegex retrieves or compiles a regex pattern.
// it uses a cache to store compiled regex patterns for reuse.
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	re, exists := regexCache[pattern]
	if exists {
		return re, nil
	}

	// double-check in case another goroutine has compiled the regex
	if re, exists = regexCache[pattern]; exists {
		return re, nil
	}

	compiledPattern := regexp.QuoteMeta(pattern)                       // Escape all regex meta characters
	compiledPattern = strings.ReplaceAll(compiledPattern, "\\*", ".*") // Replace escaped `*` with `.*` to match any character
	re, err := regexp.Compile(compiledPattern)                         // Compile the converted pattern
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

// for display purpose, replace address string with `main.xxx` when printing type.
// ref: https://github.com/gnolang/gno/pull/2357#discussion_r1704398563
func replacePackagePath(input string) string {
	result := addrRegex.ReplaceAllStringFunc(input, func(match string) string {
		parts := strings.Split(match, "/")
		if len(parts) < 4 {
			return match
		}
		lastPart := parts[len(parts)-1]
		subParts := strings.Split(lastPart, ".")
		if len(subParts) < 2 {
			return "main." + lastPart
		}
		return "main." + subParts[len(subParts)-1]
	})

	return result
}

// GetStdlibsDir returns the path to the standard libraries directory.
func GetStdlibsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot get current file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "stdlibs")
}

// hashCodeBlock generates a SHA256 hash for the given code block.
func hashCodeBlock(c codeBlock) string {
	h := sha256.New()
	h.Write([]byte(c.content))
	return hex.EncodeToString(h.Sum(nil))
}
