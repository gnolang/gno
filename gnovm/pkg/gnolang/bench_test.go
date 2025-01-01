package gnolang

import (
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
