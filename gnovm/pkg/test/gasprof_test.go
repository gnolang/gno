package test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gasprof"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/require"
)

// Phase 2: gasProfileStoreMeters routes store I/O and amino gas into the
// profiler's store dimension when profiling, and returns (nil, nil) otherwise
// so normal gno test charges no store gas.
func TestGasProfileStoreMeters(t *testing.T) {
	// Off: no store metering.
	off := &TestOptions{}
	gctx, amino := off.gasProfileStoreMeters()
	require.Nil(t, gctx)
	require.Nil(t, amino)

	// On: store I/O (via the GasContext) and amino gas both land on the store
	// dimension, attributed to the current cursor.
	p := gasprof.New()
	on := &TestOptions{GasProfiler: p}
	gctx, amino = on.gasProfileStoreMeters()
	require.NotNil(t, gctx)
	require.NotNil(t, amino)

	p.Enter(gasprof.Frame{Func: "pkg.F"})
	gctx.ConsumeGas(59_000, "DepthReadFlat")      // store read
	gctx.ConsumeGas(24_000, "DepthSet")           // store write
	amino.ConsumeGas(1_200, "AminoEncodePerByte") // amino encode

	tot := p.Totals()
	require.Equal(t, int64(59_000+24_000+1_200), tot.Store, "store + amino gas in the store dimension")
	require.Zero(t, tot.CPU)
	require.Zero(t, tot.Alloc)
}

// End-to-end: a real profiled test run must actually drive store gas through
// the wired GasContext (the unit test above only exercises the helper in
// isolation). Store gas on the dev surface is dominated by package load.
func TestGasProfile_realRunChargesStoreGas(t *testing.T) {
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)

	dir := t.TempDir()
	write := func(name, body string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
	}
	write("gnomod.toml", "module = \"gno.land/r/demo/counter\"\ngno = \"0.9\"\n")
	write("counter.gno", "package counter\n\nvar count int\n\nfunc Inc() { count++ }\nfunc Get() int { return count }\n")
	write("counter_test.gno", "package counter\n\nimport \"testing\"\n\nfunc TestInc(t *testing.T) {\n\tInc()\n\tif Get() != 1 {\n\t\tt.Fatal(\"bad\")\n\t}\n}\n")

	mpkg := gno.MustReadMemPackage(dir, "gno.land/r/demo/counter", gno.MPUserAll)

	var out bytes.Buffer
	opts := NewTestOptions(rootDir, &out, &out, nil)
	p := gasprof.New()
	opts.GasProfiler = p
	require.NoError(t, Test(mpkg, dir, opts), out.String())

	// The store dimension is now non-zero (package load + realm persistence),
	// proving store gas flows through gasProfileStoreMeters end to end. cpu is
	// also captured; without the profiler the store dimension would be zero.
	require.Positive(t, p.Totals().Store, "profiled run must charge store gas")
	require.Positive(t, p.Totals().CPU)
}

// Filetest surface: profiling a filetest now captures store gas too. Before the
// fix, filetests wrapped the machine meter after the store already held the raw
// meter, so store I/O (nil GasContext) and amino gas bypassed the profiler.
func TestGasProfile_filetestChargesStoreGas(t *testing.T) {
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)

	opts := NewTestOptions(rootDir, os.Stderr, os.Stderr, nil)
	p := gasprof.New()
	opts.GasProfiler = p

	// A realm filetest that mutates persistent state: finalizing the realm
	// forces store WRITES, whose gas is deterministic and independent of cache
	// warmth. (A bare `package main` filetest only charges store READS for
	// stdlib/type loading, which a warm object cache can drive to zero — see
	// the storage gas cold/warm trap.) This keeps the gap-fix assertion robust
	// whether the test runs first or after a warmed package cache.
	source := `// PKGPATH: gno.land/r/demo/gasproftest
package gasproftest

var saved [][]int

func main() {
	row := make([]int, 0, 8)
	for i := 0; i < 8; i++ {
		row = append(row, i*i)
	}
	saved = append(saved, row)
	println(len(saved), len(row))
}

// Output:
// 1 8
`
	report, _, _, err := opts.RunFiletest("store_filetest.gno", []byte(source), opts.TestStore)
	require.NoError(t, err, report)

	tot := p.Totals()
	require.Positive(t, tot.CPU, "cpu gas captured")
	require.Positive(t, tot.Store, "filetest store gas now captured (the gap fix)")
}
