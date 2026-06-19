package coverage_test

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test/coverage"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

const benchmarkCode = `
package main

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func sumRange(start, end int) int {
	sum := 0
	for i := start; i <= end; i++ {
		sum += i
	}
	return sum
}

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	i := 5
	for i*i <= n {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
		i += 6
	}
	return true
}

func main() {
	// Fibonacci tests
	for i := 0; i < 10; i++ {
		_ = fibonacci(i)
	}
	
	// Factorial tests
	for i := 0; i < 10; i++ {
		_ = factorial(i)
	}
	
	// Sum range tests
	_ = sumRange(1, 100)
	_ = sumRange(50, 150)
	
	// Prime number tests
	for i := 1; i <= 100; i++ {
		_ = isPrime(i)
	}
}
`

// setupMachine creates a new machine for benchmarking
func setupMachine() *gnolang.Machine {
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := gnolang.NewStore(nil, baseStore, iavlStore)
	return gnolang.NewMachine("gno.land/p/demo/main", store)
}

// setupMachineWithCoverage creates a new machine with coverage tracking enabled
func setupMachineWithCoverage() (*gnolang.Machine, *coverage.Tracker) {
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := gnolang.NewStore(nil, baseStore, iavlStore)

	tracker := coverage.NewTracker()
	tracker.SetEnabled(true)

	opts := gnolang.MachineOptions{
		PkgPath:         "gno.land/p/demo/main",
		Store:           store,
		CoverageTracker: tracker,
	}

	return gnolang.NewMachineWithOptions(opts), tracker
}

func BenchmarkWithoutCoverage(b *testing.B) {
	memPackage := &std.MemPackage{
		Type: gnolang.MPUserProd,
		Name: "main",
		Path: "gno.land/p/demo/main",
		Files: []*std.MemFile{
			{
				Name: "main.gno",
				Body: benchmarkCode,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := setupMachine()
		m.RunMemPackage(memPackage, true)
		m.RunMain()
		m.Release()
	}
}

func BenchmarkWithCoverage(b *testing.B) {
	memPackage := &std.MemPackage{
		Type: gnolang.MPUserProd,
		Name: "main",
		Path: "gno.land/p/demo/main",
		Files: []*std.MemFile{
			{
				Name: "main.gno",
				Body: benchmarkCode,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, _ := setupMachineWithCoverage()
		m.RunMemPackage(memPackage, true)
		m.RunMain()
		m.Release()
	}
}

func BenchmarkWithCoverageAndReport(b *testing.B) {
	memPackage := &std.MemPackage{
		Type: gnolang.MPUserProd,
		Name: "main",
		Path: "gno.land/p/demo/main",
		Files: []*std.MemFile{
			{
				Name: "main.gno",
				Body: benchmarkCode,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, tracker := setupMachineWithCoverage()
		m.RunMemPackage(memPackage, true)
		m.RunMain()

		_ = tracker.GenerateReport()

		m.Release()
	}
}

// Comparison benchmarks for different workloads

func BenchmarkSimpleLoopWithoutCoverage(b *testing.B) {
	simpleCode := `
package main

func main() {
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += i
	}
}
`
	pkg := &std.MemPackage{
		Type:  gnolang.MPUserProd,
		Name:  "main",
		Path:  "gno.land/p/demo/main",
		Files: []*std.MemFile{{Name: "main.gno", Body: simpleCode}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := setupMachine()
		m.RunMemPackage(pkg, true)
		m.RunMain()
		m.Release()
	}
}

func BenchmarkSimpleLoopWithCoverage(b *testing.B) {
	simpleCode := `
package main

func main() {
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += i
	}
}
`
	pkg := &std.MemPackage{
		Type:  gnolang.MPUserProd,
		Name:  "main",
		Path:  "gno.land/p/demo/main",
		Files: []*std.MemFile{{Name: "main.gno", Body: simpleCode}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, _ := setupMachineWithCoverage()
		m.RunMemPackage(pkg, true)
		m.RunMain()
		m.Release()
	}
}

func BenchmarkRecursiveWithoutCoverage(b *testing.B) {
	recursiveCode := `
package main

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
	_ = fibonacci(15)
}
`
	pkg := &std.MemPackage{
		Type:  gnolang.MPUserProd,
		Name:  "main",
		Path:  "gno.land/p/demo/main",
		Files: []*std.MemFile{{Name: "main.gno", Body: recursiveCode}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := setupMachine()
		m.RunMemPackage(pkg, true)
		m.RunMain()
		m.Release()
	}
}

func BenchmarkRecursiveWithCoverage(b *testing.B) {
	recursiveCode := `
package main

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
	_ = fibonacci(15)
}
`
	pkg := &std.MemPackage{
		Type:  gnolang.MPUserProd,
		Name:  "main",
		Path:  "gno.land/p/demo/main",
		Files: []*std.MemFile{{Name: "main.gno", Body: recursiveCode}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, _ := setupMachineWithCoverage()
		m.RunMemPackage(pkg, true)
		m.RunMain()
		m.Release()
	}
}
