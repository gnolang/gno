package gnolang

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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
