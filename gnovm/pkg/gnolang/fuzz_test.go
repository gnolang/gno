package gnolang

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cockroachdb/apd/v3"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
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
	parseFileDir := filepath.Join("testdata", "corpora", "parsefile")
	paths, err := filepath.Glob(filepath.Join(parseFileDir, "*.go_fuzz"))
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

type convertToCorpus struct {
	TV  *TypedValue `json:"tv"`
	Typ Type        `json:"typ"`
}

func mustDecFromStr(s string) *apd.Decimal {
	dec, _, err := apd.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return dec
}

func FuzzConvertTo(f *testing.F) {
	seeds := []*convertToCorpus{
		{
			TV: &TypedValue{
				T: Float32Type,
				V: &BigdecValue{V: mustDecFromStr("10.999")},
			},
			Typ: Float64Type,
		},
		{
			TV: &TypedValue{
				T: Float64Type,
				V: &BigdecValue{V: mustDecFromStr("10.999")},
			},
			Typ: Float64Type,
		},
		{
			TV: &TypedValue{
				T: Float64Type,
				V: &BigdecValue{V: mustDecFromStr("10.999")},
			},
			Typ: Float32Type,
		},
		{
			TV: &TypedValue{
				T: Int64Type,
				V: &BigintValue{V: big.NewInt(-9816)},
			},
			Typ: Float32Type,
		},
		{
			TV: &TypedValue{
				T: Int64Type,
				V: &BigintValue{V: big.NewInt(9816)},
			},
			Typ: Float32Type,
		},
		{
			TV: &TypedValue{
				T: Uint64Type,
				V: &BigintValue{V: big.NewInt(9816)},
			},
			Typ: Float32Type,
		},
		{
			TV: &TypedValue{
				T: Uint64Type,
				V: &BigintValue{V: big.NewInt(9816)},
			},
			Typ: Float64Type,
		},
	}

	for _, seed := range seeds {
		blob, err := json.Marshal(seed)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(blob)
	}

	f.Fuzz(func(t *testing.T, inputJSON []byte) {
		cc := new(convertToCorpus)
		if err := json.Unmarshal(inputJSON, cc); err != nil {
			return
		}
		if cc.TV == nil || cc.TV.V == nil {
			return
		}

		db := memdb.NewMemDB()
		baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
		iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
		store := NewStore(nil, baseStore, iavlStore)
		allocator := NewAllocator(1 << 20)
		ConvertTo(allocator, store, cc.TV, cc.Typ, true)
		ConvertTo(allocator, store, cc.TV, cc.Typ, false)
	})
}
