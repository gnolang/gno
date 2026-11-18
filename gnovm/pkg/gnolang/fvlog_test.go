package gnolang

import (
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// fvReset closes the probe's file and lets the next fvOn re-read the env.
func fvReset(t *testing.T) {
	t.Helper()
	if fvFile != nil {
		fvFile.Close()
	}
	fvFile = nil
	fvOnce = sync.Once{}
}

// fvForeignBirth makes a struct of a type realm a declares while realm b is current.
func fvForeignBirth() {
	dt := &DeclaredType{PkgPath: "gno.land/r/a", Name: "T", Base: &StructType{PkgPath: "gno.land/r/a"}}
	alloc := NewAllocator(math.MaxInt64)
	alloc.currentRealmPath = "gno.land/r/b"
	alloc.NewStruct(dt, nil)
}

func TestFvLogOff(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fv.log")
	t.Setenv("GNO_FV_LOG", "")
	fvReset(t)
	t.Cleanup(func() { fvReset(t) })

	fvForeignBirth()
	if fvFile != nil {
		t.Fatal("probe opened a file with GNO_FV_LOG unset")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("log file exists: %v", err)
	}
}

func TestFvLogOn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fv.log")
	t.Setenv("GNO_FV_LOG", path)
	fvReset(t)
	t.Cleanup(func() { fvReset(t) })

	fvForeignBirth()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "fv birth gno.land/r/a.T in gno.land/r/b code=- caller=- at NewStruct\n"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
