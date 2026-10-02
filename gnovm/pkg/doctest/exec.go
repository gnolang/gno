package doctest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/store"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	gnoLang = "gno"
	// gnoDoctest is an alternate tag so authors can write "go,gnodoctest"
	// and still get GitHub syntax highlighting for Gno code.
	gnoDoctest = "gnodoctest"

	mainPkgPath   = "main"
	maxAllocBytes = 500_000_000
)

func isGnoDoctest(lang string) bool {
	for part := range strings.SplitSeq(lang, ",") {
		if strings.TrimSpace(part) == gnoDoctest {
			return true
		}
	}
	return false
}

// ExecuteCodeBlock runs a single block. Each call rebuilds the
// stdlib store; use [ExecuteMatchingCodeBlock] to amortize that cost
// across multiple blocks.
func ExecuteCodeBlock(c codeBlock, rootDir string) (string, error) {
	if c.options.Ignore {
		return "IGNORED", nil
	}
	if skipped, ok := unsupportedLangResult(c.lang); ok {
		return skipped, nil
	}
	baseStore, gnoStore := test.ProdStore(rootDir, io.Discard, nil)
	return executeBlock(c, baseStore, gnoStore)
}

// executeBlock runs one block against a prepared store. The store is
// cache-wrapped per call so blocks sharing gnoStore stay isolated.
func executeBlock(c codeBlock, baseStore storetypes.CommitStore, gnoStore gno.Store) (string, error) {
	output, runErr := runGnoBlock(c, baseStore, gnoStore)

	if c.options.ShouldPanic {
		return handlePanicMessage(runErr, c.options.PanicMessage)
	}
	if runErr != nil {
		return "", runErr
	}
	if c.expectedOutput == "" && c.expectedError == "" {
		return output, nil
	}
	return compareResults(output, c.expectedOutput, c.expectedError)
}

func runGnoBlock(c codeBlock, baseStore storetypes.CommitStore, gnoStore gno.Store) (_ string, err error) {
	buf := new(bytes.Buffer)
	gasMeter := store.NewInfiniteGasMeter()
	tcw := baseStore.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:       mainPkgPath,
		Output:        buf,
		Store:         gnoStore.BeginTransaction(tcw, tcw, gasMeter),
		Context:       test.Context(test.DefaultCaller, mainPkgPath, nil),
		MaxAllocBytes: maxAllocBytes,
		GasMeter:      gasMeter,
	})
	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			if upe, ok := r.(gno.UnhandledPanicError); ok {
				err = errors.New(upe.Error())
				return
			}
			err = fmt.Errorf("%v", r)
		}
	}()

	file, err := m.ParseFile(fmt.Sprintf("%d.gno", c.index), c.content)
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}
	m.RunFiles(file)

	mainExpr, err := m.ParseExpr("main()")
	if err != nil {
		return "", fmt.Errorf("parse main(): %w", err)
	}
	m.Eval(mainExpr)

	return buf.String(), nil
}

func unsupportedLangResult(lang string) (string, bool) {
	base := strings.Split(lang, ",")[0]
	if base == gnoLang || isGnoDoctest(lang) {
		return "", false
	}
	return fmt.Sprintf("SKIPPED (Unsupported language: %s)", base), true
}

// ExecuteMatchingCodeBlock runs every code block whose name matches
// pattern, sharing one stdlib store across them.
func ExecuteMatchingCodeBlock(
	ctx context.Context,
	content string,
	pattern string,
	rootDir string,
) ([]string, error) {
	codeBlocks, err := GetCodeBlocks(content)
	if err != nil {
		return nil, err
	}

	var (
		baseStore storetypes.CommitStore
		gnoStore  gno.Store
	)
	results := make([]string, 0, len(codeBlocks))
	for _, block := range codeBlocks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if !matchPattern(block.name, pattern) {
			continue
		}

		if block.options.Ignore {
			results = append(results, fmt.Sprintf("\n=== %s ===\n\nIGNORED\n", block.name))
			continue
		}
		if skipped, ok := unsupportedLangResult(block.lang); ok {
			results = append(results, fmt.Sprintf("\n=== %s ===\n\n%s\n", block.name, skipped))
			continue
		}

		if gnoStore == nil {
			baseStore, gnoStore = test.ProdStore(rootDir, io.Discard, nil)
		}

		result, err := executeBlock(block, baseStore, gnoStore)
		if err != nil {
			return nil, fmt.Errorf("failed to execute code block %s: %w", block.name, err)
		}
		results = append(results, fmt.Sprintf("\n=== %s ===\n\n%s\n", block.name, result))
	}

	return results, nil
}

// handlePanicMessage validates a should_panic block. An empty
// panicMessage matches any panic.
func handlePanicMessage(err error, panicMessage string) (string, error) {
	if err == nil {
		if panicMessage == "" {
			return "", errors.New("expected a panic, but executed successfully")
		}
		return "", fmt.Errorf("expected panic with message: %s, but executed successfully", panicMessage)
	}
	if panicMessage == "" || strings.Contains(err.Error(), panicMessage) {
		return fmt.Sprintf("panicked as expected: %v", err), nil
	}
	return "", fmt.Errorf("expected panic with message: %s, but got: %s", panicMessage, err.Error())
}

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

	if pattern, ok := strings.CutPrefix(expected, "regex:"); ok {
		return compareRegex(actual, pattern)
	}

	if actual != expected {
		return "", fmt.Errorf("expected:\n%s\n\nbut got:\n%s", expected, actual)
	}

	return actual, nil
}

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

// matchPattern treats pattern as a regexp (like `go test -run`);
// an empty pattern matches everything.
func matchPattern(name, pattern string) bool {
	if pattern == "" {
		return true
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(name)
}

// DefaultRootDir returns the gno root directory used by doctest when
// no rootDir is supplied.
func DefaultRootDir() string {
	return gnoenv.RootDir()
}
