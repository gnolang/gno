package gnolang

import (
	"archive/zip"
	"io"
	"net/http"
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

	// 1.5. Next we should get more valid and diverse Go files.
	// Opt-ing to always download from the latest Go tree on Github instead
	// of committing 35+MiB into this Git tree's history for life.
	goTreeURL := "https://github.com/golang/go/archive/refs/heads/master.zip"
	res, err := http.Get(goTreeURL)
	if err != nil {
		f.Fatal(err)
	}
	defer res.Body.Close()
	// Write the downloaded zip into the temporary directory so that
	// zip.OpenReader can use an io.ReaderAt instead of the io.ReadCloser
	// that http.Response.Body is.
	goTreeZipLocally := filepath.Join(f.TempDir(), "go-tree.zip")
	fz, err := os.Create(goTreeZipLocally)
	if err != nil {
		f.Fatal(err)
	}
	_, _ = io.Copy(fz, res.Body)
	res.Body.Close()
	fz.Close()

	goZip, err := zip.OpenReader(goTreeZipLocally)
	if err != nil {
		f.Fatal(err)
	}
	for _, fz := range goZip.File {
		if !strings.HasSuffix(fz.Name, ".go") {
			continue
		}

		// We only want to add files in the "/test/" directory or *_test.go
		// for valid and diverse Go code samples.
		acceptable := strings.Contains(fz.Name, "/test/") || strings.HasSuffix(fz.Name, "test.go")
		if !acceptable {
			continue
		}

		fi := fz.FileInfo()
		if fi.IsDir() {
			continue
		}

		rc, err := fz.Open()
		if err != nil {
			f.Fatal(err)
		}
		goProgramBody, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			f.Fatal(err)
		}

		f.Add(string(goProgramBody))
	}

	// 2. Now run the fuzzer.
	f.Fuzz(func(t *testing.T, goFileContents string) {
		_, _ = ParseFile("a.go", goFileContents)
	})
}
