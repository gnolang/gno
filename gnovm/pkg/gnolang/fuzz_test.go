package gnolang

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
)

func FuzzConvertUntypedBigdecToFloat(f *testing.F) {
	// 1. Firstly add seeds.
	seeds := []string{
		"-100000",
		"100000",
		"0",
	}

	check := new(apd.Decimal)
	for _, seed := range seeds {
		if check.UnmarshalText([]byte(seed)) == nil {
			f.Add(seed)
		}
	}

	f.Fuzz(func(t *testing.T, apdStr string) {
		switch {
		case strings.HasPrefix(apdStr, ".-"):
			return
		}

		v := new(apd.Decimal)
		if err := v.UnmarshalText([]byte(apdStr)); err != nil {
			return
		}
		if _, err := v.Float64(); err != nil {
			return
		}

		bd := BigdecValue{
			V: v,
		}
		dst := new(TypedValue)
		typ := Float64Type
		ConvertUntypedBigdecTo(dst, bd, typ)
	})
}

func FuzzParseFile(f *testing.F) {
	// 1. Add the corpra.
	parseFileDir := filepath.Join("testdata", "corpra", "parsefile")
	paths, err := filepath.Glob(filepath.Join(parseFileDir, "*.go"))
	if err != nil {
		f.Fatal(err)
	}

	// Also load in files from gno/gnovm/tests/files
	pc, curFile, _, _ := runtime.Caller(0)
	curFileDir := filepath.Dir(curFile)
	gnovmTestFilesDir, err := filepath.Abs(filepath.Join(curFileDir, "..", "..", "tests", "files"))
	if err != nil {
		_ = pc // To silence the arbitrary golangci linter.
		f.Fatal(err)
	}
	globGnoTestFiles := filepath.Join(gnovmTestFilesDir, "*.gno")
	gnoTestFiles, err := filepath.Glob(globGnoTestFiles)
	if err != nil {
		f.Fatal(err)
	}
	if len(gnoTestFiles) == 0 {
		f.Fatalf("no files found from globbing %q", globGnoTestFiles)
	}
	paths = append(paths, gnoTestFiles...)

	for _, path := range paths {
		blob, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(string(blob))
	}

	// 2. Now run the fuzzer.
	f.Fuzz(func(t *testing.T, goFileContents string) {
		_, _ = ParseFile("a.go", goFileContents)
	})
}

func FuzzDoOpEvalBaseConversion(f *testing.F) {
	if testing.Short() {
		f.Skip("Skipping in -short mode")
	}

	// 1. Add the seeds.
	seeds := []*BasicLitExpr{
		{Kind: INT, Value: "9223372036854775807"},
		{Kind: INT, Value: "0"},
		{Kind: INT, Value: "0777"},
		{Kind: INT, Value: "0xDEADBEEF"},
		{Kind: INT, Value: "0o123"},
		{Kind: INT, Value: "0b111111111111111111111111111111111111111111111111111111111111111"},
		{Kind: FLOAT, Value: "0.00001111"},
		{Kind: FLOAT, Value: "9999.12"},
		{Kind: STRING, Value: `"9999.12"`},
		{Kind: STRING, Value: `"aaaaaaaaaaa   "`},
		{Kind: STRING, Value: `"🚨🌎"`},
	}

	for _, seed := range seeds {
		blob, err := json.Marshal(seed)
		if err != nil {
			panic(err)
		}
		f.Add(blob)
	}

	// 2. Fuzz it.
	f.Fuzz(func(t *testing.T, basicLitExprBlob []byte) {
		expr := new(BasicLitExpr)
		if err := json.Unmarshal(basicLitExprBlob, expr); err != nil {
			return
		}

		m := NewMachine("test", nil)

		defer func() {
			r := recover()
			if r == nil {
				return
			}

			s := fmt.Sprintf("%s", r)
			switch {
			case strings.Contains(s, "unexpected lit kind"):
				return

			case strings.Contains(s, "unexpected decimal/float format"):
				if strings.Contains(s, "/") {
					return
				}
				fallthrough
			case strings.Contains(s, "invalid integer constant"):
				if containsAny(s, "+", "-", "*", "/", "%", "&", "|", "||") {
					return
				}
				fallthrough
			default:
				if !basicLitExprIsValidGoValue(t, expr) {
					return
				}

				if expr.Kind != STRING {
					trimmed := strings.TrimSpace(expr.Value)
					if trimmed != expr.Value {
						expr.Value = trimmed
						m := NewMachine("test", nil)
						m.PushExpr(expr)
						m.doOpEval()
						_ = m.PopValue()
						return
					}
				}

				panic(r)
			}
		}()

		m.PushExpr(expr)
		m.doOpEval()
		_ = m.PopValue()
	})
}

func containsAny(s string, symbols ...string) bool {
	for _, sym := range symbols {
		if strings.Contains(s, sym) {
			return true
		}
	}
	return false
}

func basicLitExprIsValidGoValue(t *testing.T, lit *BasicLitExpr) bool {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "basic_lit_check.go")
	var craftedGo = []byte(fmt.Sprintf(`package main
const _ %s = %s
func main() {}`, strings.ToLower(lit.Kind.String()), lit.Value))
	if err := os.WriteFile(file, craftedGo, 0755); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, "go", "run", file).Run()
	return err == nil
}
