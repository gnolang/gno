package test

import (
	"fmt"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/require"
)

// benchHeadFmt builds the cycles/gas prefix using the production column
// widths so tests fail loudly if those widths intentionally change.
func benchHeadFmt(nameWidth int) string {
	return fmt.Sprintf("%%-%ds\t%%%dd\t%%%dd cycles/op\t%%%dd gas/op",
		nameWidth, benchColN, benchColCycles, benchColGas)
}

func TestFormatBenchmarkResult(t *testing.T) {
	t.Parallel()

	const name = "BenchmarkFoo"
	head := benchHeadFmt(len(name))
	bytesFmt := fmt.Sprintf("\t%%%dd bytes/op", benchColBytes)
	memFmt := fmt.Sprintf("\t%%%dd B/op\t%%%dd allocs/op", benchColAllocBytes, benchColAllocCount)

	cases := []struct {
		desc     string
		rep      benchmarkReport
		benchmem bool
		want     string
	}{
		{
			desc:     "cycles_only",
			rep:      benchmarkReport{N: 10, Cycles: 4200},
			benchmem: false,
			want:     fmt.Sprintf(head, name, 10, 420, 0),
		},
		{
			desc:     "gas_appears_alongside_cycles",
			rep:      benchmarkReport{N: 10, Cycles: 4200, Gas: 5000},
			benchmem: false,
			want:     fmt.Sprintf(head, name, 10, 420, 500),
		},
		{
			desc:     "set_bytes_adds_bytes_per_op",
			rep:      benchmarkReport{N: 10, Cycles: 4200, Gas: 5000, Bytes: 7},
			benchmem: false,
			want:     fmt.Sprintf(head+bytesFmt, name, 10, 420, 500, 7),
		},
		{
			desc:     "benchmem_flag_enables_mem_columns",
			rep:      benchmarkReport{N: 10, Cycles: 4200, Gas: 5000, AllocBytes: 1000, Allocs: 50},
			benchmem: true,
			want:     fmt.Sprintf(head+memFmt, name, 10, 420, 500, 100, 5),
		},
		{
			desc:     "reportallocs_enables_mem_columns_without_flag",
			rep:      benchmarkReport{N: 10, Cycles: 4200, Gas: 5000, AllocBytes: 1000, Allocs: 50, ReportAllocs: true},
			benchmem: false,
			want:     fmt.Sprintf(head+memFmt, name, 10, 420, 500, 100, 5),
		},
		{
			desc:     "benchmem_and_reportallocs_do_not_double_print",
			rep:      benchmarkReport{N: 10, Cycles: 4200, Gas: 5000, AllocBytes: 1000, Allocs: 50, ReportAllocs: true},
			benchmem: true,
			want:     fmt.Sprintf(head+memFmt, name, 10, 420, 500, 100, 5),
		},
		{
			desc:     "bytes_and_benchmem_both_appear",
			rep:      benchmarkReport{N: 10, Cycles: 4200, Gas: 5000, Bytes: 7, AllocBytes: 1000, Allocs: 50},
			benchmem: true,
			want:     fmt.Sprintf(head+bytesFmt+memFmt, name, 10, 420, 500, 7, 100, 5),
		},
		{
			desc:     "zero_n_uses_one_as_divisor",
			rep:      benchmarkReport{N: 0, Cycles: 4200, Gas: 5000, AllocBytes: 1000, Allocs: 50},
			benchmem: true,
			want:     fmt.Sprintf(head+memFmt, name, 0, 4200, 5000, 1000, 50),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			got := formatBenchmarkResult(name, tc.rep, tc.benchmem, len(name))
			require.Equal(t, tc.want, got)
		})
	}
}

func TestFormatBenchmarkResultNamePadding(t *testing.T) {
	t.Parallel()

	got := formatBenchmarkResult("BenchA", benchmarkReport{N: 1, Cycles: 100, Gas: 200}, false, 12)
	want := fmt.Sprintf(benchHeadFmt(12), "BenchA", 1, 100, 200)
	require.Equal(t, want, got)

	// nameWidth smaller than name length must not truncate.
	got = formatBenchmarkResult("BenchmarkLongName", benchmarkReport{N: 1, Cycles: 100, Gas: 200}, false, 4)
	want = fmt.Sprintf(benchHeadFmt(len("BenchmarkLongName")), "BenchmarkLongName", 1, 100, 200)
	require.Equal(t, want, got)
}

func TestLoadBenchFuncs(t *testing.T) {
	t.Parallel()

	src := `package bench

import "testing"

func BenchmarkOne(b *testing.B) {}
func BenchmarkTwo(b *testing.B) {}
func BenchmarkCrossing(cur realm, b *testing.B) {}
func TestSomething(t *testing.T) {}
func FuzzSomething(f *testing.F) {}
func ExampleSomething() {}
func helper() {}

type S struct{}

func (s *S) BenchmarkMethod(b *testing.B) {}
`

	var m *gno.Machine // ParseFile tolerates a nil receiver.
	fn, err := m.ParseFile("bench_test.gno", src)
	require.NoError(t, err)

	fset := &gno.FileSet{}
	fset.AddFiles(fn)

	got := loadBenchFuncs("bench", fset)
	names := make([]string, len(got))
	for i, tf := range got {
		names[i] = tf.Name
	}
	require.ElementsMatch(t,
		[]string{"BenchmarkOne", "BenchmarkTwo", "BenchmarkCrossing"},
		names,
		"loadBenchFuncs should collect only top-level functions with the Benchmark prefix",
	)
}

func TestShouldRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		desc    string
		pattern string
		input   string
		want    bool
	}{
		{"nil_filter_matches_all", "", "BenchmarkAny", true},
		{"exact_substring_matches", "BenchmarkFoo", "BenchmarkFoo", true},
		{"partial_substring_matches", "Foo", "BenchmarkFoo", true},
		{"nonexistent_pattern_does_not_match", "Nonexistent", "BenchmarkFoo", false},
		{"anchored_end_matches_exact_suffix", "Foo$", "BenchmarkFoo", true},
		{"anchored_end_rejects_extra_suffix", "Foo$", "BenchmarkFooBar", false},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			var filter filterMatch
			if tc.pattern != "" {
				filter = splitRegexp(tc.pattern)
			}
			require.Equal(t, tc.want, shouldRun(filter, tc.input))
		})
	}
}
