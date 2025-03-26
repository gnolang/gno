package gnolang

import (
	"path/filepath"
	"runtime"
	"testing"
)

var sink any = nil

var pkgIDPaths = []string{
	"encoding/json",
	"math/bits",
	"github.com/gnolang/gno/gnovm/pkg/gnolang",
	"a",
	" ",
	"",
	"github.com/gnolang/gno/gnovm/pkg/gnolang/vendor/pkg/github.com/gnolang/vendored",
}

func BenchmarkPkgIDFromPkgPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, path := range pkgIDPaths {
			sink = PkgIDFromPkgPath(path)
		}
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}
	sink = nil
}

func BenchmarkReadMemPackage(b *testing.B) {
	_, currentFile, _, ok := runtime.Caller(0)
	_ = ok // Appease golang-ci
	rootOfRepo, err := filepath.Abs(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	if err != nil {
		b.Fatal(err)
	}
	demoDir := filepath.Join(rootOfRepo, "examples", "gno.land", "p", "demo")
	ufmtDir := filepath.Join(demoDir, "ufmt")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink = MustReadMemPackage(ufmtDir, "ufmt")
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}
	sink = nil
}
