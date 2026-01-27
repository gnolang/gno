//go:build gnobench

package benchops

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWritePprof(t *testing.T) {
	p := New()
	p.Start()

	// Add some operations with location context
	p.BeginOp(OpAdd)
	p.SetOpContext(OpContext{
		File:     "test.gno",
		Line:     10,
		FuncName: "add",
		PkgPath:  "gno.land/r/demo/test",
	})
	time.Sleep(time.Microsecond)
	p.EndOp()

	p.BeginOp(OpMul)
	p.SetOpContext(OpContext{
		File:     "test.gno",
		Line:     20,
		FuncName: "mul",
		PkgPath:  "gno.land/r/demo/test",
	})
	time.Sleep(time.Microsecond)
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WritePprof(&buf)
	require.NoError(t, err)
	assert.NotZero(t, buf.Len(), "expected non-empty pprof output")

	// Parse using the pprof library to validate format
	prof, err := profile.Parse(&buf)
	require.NoError(t, err, "pprof output should be valid profile format")

	// Verify profile structure
	require.Len(t, prof.SampleType, 1, "expected 1 sample type")
	assert.Equal(t, "gas", prof.SampleType[0].Type)
	assert.Equal(t, "units", prof.SampleType[0].Unit)

	// Verify samples exist
	require.Len(t, prof.Sample, 2, "expected 2 samples (one per location)")

	// Verify locations and functions
	require.Len(t, prof.Location, 2)
	require.Len(t, prof.Function, 2)

	// Verify function names are present
	funcNames := make(map[string]bool)
	for _, fn := range prof.Function {
		funcNames[fn.Name] = true
	}
	assert.Contains(t, funcNames, "add")
	assert.Contains(t, funcNames, "mul")
}

func TestWritePprofNilResults(t *testing.T) {
	var r *Results

	var buf bytes.Buffer
	err := r.WritePprof(&buf)
	require.NoError(t, err)
	assert.Zero(t, buf.Len(), "expected empty output for nil results")
}

func TestWritePprofEmptyResults(t *testing.T) {
	p := New()
	p.Start()
	results := p.Stop() // No operations recorded

	var buf bytes.Buffer
	err := results.WritePprof(&buf)
	require.NoError(t, err)

	// Empty results should produce valid (but minimal) pprof
	if buf.Len() > 0 {
		prof, err := profile.Parse(&buf)
		require.NoError(t, err)
		assert.Empty(t, prof.Sample, "empty results should have no samples")
	}
}

func TestWritePprofGasValues(t *testing.T) {
	p := New()
	p.Start()

	// Add 3 ops at same location to accumulate gas
	for i := 0; i < 3; i++ {
		p.BeginOp(OpAdd) // 18 gas each
		p.SetOpContext(OpContext{
			File:     "test.gno",
			Line:     10,
			FuncName: "add",
			PkgPath:  "gno.land/r/demo/test",
		})
		p.EndOp()
	}

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WritePprof(&buf)
	require.NoError(t, err)

	prof, err := profile.Parse(&buf)
	require.NoError(t, err)

	// Verify gas value (3 * 18 = 54)
	require.Len(t, prof.Sample, 1)
	assert.Equal(t, int64(54), prof.Sample[0].Value[0])
}

func TestWritePprofLineNumbers(t *testing.T) {
	p := New()
	p.Start()

	p.BeginOp(OpAdd)
	p.SetOpContext(OpContext{
		File:     "test.gno",
		Line:     42,
		FuncName: "myFunc",
		PkgPath:  "gno.land/r/demo/test",
	})
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WritePprof(&buf)
	require.NoError(t, err)

	prof, err := profile.Parse(&buf)
	require.NoError(t, err)

	// Verify line number is correct
	require.Len(t, prof.Location, 1)
	require.Len(t, prof.Location[0].Line, 1)
	assert.Equal(t, int64(42), prof.Location[0].Line[0].Line)

	// Verify function start line
	require.Len(t, prof.Function, 1)
	assert.Equal(t, int64(42), prof.Function[0].StartLine)
}

func TestWritePprofWithStacks(t *testing.T) {
	// Build results with pre-aggregated stack samples
	// (aggregation happens in buildResults(), not in pprof builder)
	results := &Results{
		StackSamples: []*StackSample{
			{
				Stack: []StackFrame{
					{Func: "leaf", File: "a.gno", Line: 10},
					{Func: "middle", File: "b.gno", Line: 20},
					{Func: "root", File: "c.gno", Line: 30},
				},
				Gas:   150, // pre-aggregated: 100 + 50
				Count: 2,
			},
		},
	}

	var buf bytes.Buffer
	err := results.WritePprof(&buf)
	require.NoError(t, err)

	prof, err := profile.Parse(&buf)
	require.NoError(t, err)

	// Verify stack sample is output correctly
	require.Len(t, prof.Sample, 1)
	assert.Equal(t, int64(150), prof.Sample[0].Value[0], "gas should match")

	// Verify stack depth
	require.Len(t, prof.Sample[0].Location, 3, "should have 3 locations in stack")
}

func TestWritePprofWithOptions_MultipleSampleTypes(t *testing.T) {
	// Build results with pre-aggregated stack samples
	// (aggregation happens in buildResults(), not in pprof builder)
	results := &Results{
		TimingEnabled: true,
		StackSamples: []*StackSample{
			{
				Stack: []StackFrame{
					{Func: "leaf", File: "a.gno", Line: 10, PkgPath: "gno.land/r/demo"},
				},
				Gas:        150,  // pre-aggregated: 100 + 50
				DurationNs: 7500, // pre-aggregated: 5000 + 2500
				Count:      5,    // pre-aggregated: 3 + 2
			},
		},
	}

	var buf bytes.Buffer
	err := results.WritePprofWithOptions(&buf, WithDuration(), WithCount())
	require.NoError(t, err)

	prof, err := profile.Parse(&buf)
	require.NoError(t, err)

	// Verify we have 3 sample types: gas, duration, count
	require.Len(t, prof.SampleType, 3)
	assert.Equal(t, "gas", prof.SampleType[0].Type)
	assert.Equal(t, "units", prof.SampleType[0].Unit)
	assert.Equal(t, "duration", prof.SampleType[1].Type)
	assert.Equal(t, "nanoseconds", prof.SampleType[1].Unit)
	assert.Equal(t, "count", prof.SampleType[2].Type)
	assert.Equal(t, "samples", prof.SampleType[2].Unit)

	// Verify values are output correctly
	require.Len(t, prof.Sample, 1)
	require.Len(t, prof.Sample[0].Value, 3)
	assert.Equal(t, int64(150), prof.Sample[0].Value[0], "gas should match")
	assert.Equal(t, int64(7500), prof.Sample[0].Value[1], "duration should match")
	assert.Equal(t, int64(5), prof.Sample[0].Value[2], "count should match")
}

func TestWritePprofWithOptions_Labels(t *testing.T) {
	results := &Results{
		StackSamples: []*StackSample{
			{
				Stack: []StackFrame{
					{Func: "leaf", File: "a.gno", Line: 10, PkgPath: "gno.land/r/demo"},
					{Func: "middle", File: "b.gno", Line: 20, PkgPath: "gno.land/r/demo"},
				},
				Gas:   100,
				Count: 1,
			},
		},
	}

	var buf bytes.Buffer
	err := results.WritePprofWithOptions(&buf, WithLabels())
	require.NoError(t, err)

	prof, err := profile.Parse(&buf)
	require.NoError(t, err)

	require.Len(t, prof.Sample, 1)

	// Verify labels are present
	sample := prof.Sample[0]
	assert.Contains(t, sample.Label, "pkg")
	assert.Equal(t, []string{"gno.land/r/demo"}, sample.Label["pkg"])

	// Verify numeric labels (depth)
	assert.Contains(t, sample.NumLabel, "depth")
	assert.Equal(t, []int64{2}, sample.NumLabel["depth"])
}

func TestWritePprofWithOptions_DurationDisabledWithoutTiming(t *testing.T) {
	// When timing is not enabled, duration should not be included even if requested
	results := &Results{
		TimingEnabled: false,
		StackSamples: []*StackSample{
			{
				Stack: []StackFrame{
					{Func: "leaf", File: "a.gno", Line: 10},
				},
				Gas:        100,
				DurationNs: 5000, // This should be ignored
				Count:      1,
			},
		},
	}

	var buf bytes.Buffer
	err := results.WritePprofWithOptions(&buf, WithDuration(), WithCount())
	require.NoError(t, err)

	prof, err := profile.Parse(&buf)
	require.NoError(t, err)

	// Only gas and count (no duration since timing was disabled)
	require.Len(t, prof.SampleType, 2)
	assert.Equal(t, "gas", prof.SampleType[0].Type)
	assert.Equal(t, "count", prof.SampleType[1].Type)
}

func TestWritePprofWithOptions_EnhancedMetadata(t *testing.T) {
	results := &Results{
		StackSamples: []*StackSample{
			{
				Stack: []StackFrame{
					{Func: "myFunc", File: "test.gno", Line: 42, PkgPath: "gno.land/r/demo/test"},
				},
				Gas:   100,
				Count: 1,
			},
		},
	}

	var buf bytes.Buffer
	err := results.WritePprof(&buf)
	require.NoError(t, err)

	prof, err := profile.Parse(&buf)
	require.NoError(t, err)

	// Verify system name includes package path
	require.Len(t, prof.Function, 1)
	assert.Equal(t, "myFunc", prof.Function[0].Name)
	assert.Equal(t, "gno.land/r/demo/test.myFunc", prof.Function[0].SystemName)
}

func TestWritePprofWithOptions_LocationStats(t *testing.T) {
	// Test that location stats also support multi-sample types
	results := &Results{
		TimingEnabled: true,
		LocationStats: []*LocationStat{
			{
				File:     "test.gno",
				Line:     10,
				FuncName: "foo",
				PkgPath:  "gno.land/r/demo",
				Count:    5,
				TotalNs:  10000,
				Gas:      200,
			},
		},
	}

	var buf bytes.Buffer
	err := results.WritePprofWithOptions(&buf, WithDuration(), WithCount(), WithLabels())
	require.NoError(t, err)

	prof, err := profile.Parse(&buf)
	require.NoError(t, err)

	// Verify sample types
	require.Len(t, prof.SampleType, 3)

	// Verify sample values
	require.Len(t, prof.Sample, 1)
	assert.Equal(t, int64(200), prof.Sample[0].Value[0])   // gas
	assert.Equal(t, int64(10000), prof.Sample[0].Value[1]) // duration
	assert.Equal(t, int64(5), prof.Sample[0].Value[2])     // count

	// Verify labels
	assert.Contains(t, prof.Sample[0].Label, "pkg")
	assert.Equal(t, []string{"gno.land/r/demo"}, prof.Sample[0].Label["pkg"])

	// Verify function system name includes package path
	require.Len(t, prof.Function, 1)
	assert.Equal(t, "gno.land/r/demo.foo", prof.Function[0].SystemName)
}
