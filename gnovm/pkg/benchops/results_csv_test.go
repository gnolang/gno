package benchops

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteCSV(t *testing.T) {
	p := New()
	// timingEnabled is true by default, so CSV output includes timing columns
	p.Start()

	// Add some operations
	p.BeginOp(OpAdd, OpContext{
		File:     "test.gno",
		Line:     10,
		FuncName: "add",
		PkgPath:  "gno.land/r/demo/test",
	})
	p.EndOp()

	p.BeginOp(OpMul, OpContext{
		File:     "test.gno",
		Line:     20,
		FuncName: "mul",
		PkgPath:  "gno.land/r/demo/test",
	})
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteCSV(&buf)
	require.NoError(t, err)

	csv := buf.String()

	// Check OpStats header and data (with timing columns since timing is enabled by default)
	assert.Contains(t, csv, "opcode,count,gas,total_ns,avg_ns,stddev_ns,min_ns,max_ns")
	assert.Contains(t, csv, "OpAdd,1,")
	assert.Contains(t, csv, "OpMul,1,")

	// Check LocationStats header and data
	assert.Contains(t, csv, "file,line,func,pkg,count,gas")
	assert.Contains(t, csv, "test.gno,10,add,gno.land/r/demo/test")
	assert.Contains(t, csv, "test.gno,20,mul,gno.land/r/demo/test")
}

func TestWriteCSVWithoutTiming(t *testing.T) {
	p := New()
	// Internal test: directly set field to test disabled behavior.
	// Production code uses WithoutTiming() option via Start().
	p.timingEnabled = false
	p.Start()

	p.BeginOp(OpAdd, OpContext{})
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteCSV(&buf)
	require.NoError(t, err)

	csv := buf.String()

	// Check timing columns are NOT present when timing is disabled
	assert.Contains(t, csv, "opcode,count,gas")
	assert.NotContains(t, csv, "total_ns")
}

func TestWriteCSVNilResults(t *testing.T) {
	var r *Results

	var buf bytes.Buffer
	err := r.WriteCSV(&buf)
	require.NoError(t, err)
	assert.Empty(t, buf.String())
}

func TestWriteCSVSection(t *testing.T) {
	p := New()
	// Internal test: directly set field to test disabled behavior.
	// Production code uses WithoutTiming() option via Start().
	p.timingEnabled = false
	p.Start()

	p.BeginOp(OpAdd, OpContext{})
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteCSVSection(&buf, SectionOpcodes)
	require.NoError(t, err)

	csv := buf.String()
	lines := strings.Split(strings.TrimSpace(csv), "\n")

	// Should have header + 1 data row
	require.Len(t, lines, 2)
	assert.Equal(t, "opcode,count,gas", lines[0])
}

func TestWriteCSVSectionWithTiming(t *testing.T) {
	p := New()
	// timingEnabled is true by default
	p.Start()

	p.BeginOp(OpAdd, OpContext{})
	p.EndOp()

	results := p.Stop()

	var buf bytes.Buffer
	err := results.WriteCSVSection(&buf, SectionOpcodes)
	require.NoError(t, err)

	csv := buf.String()
	lines := strings.Split(strings.TrimSpace(csv), "\n")

	// Should have header + 1 data row with timing columns
	require.Len(t, lines, 2)
	assert.Equal(t, "opcode,count,gas,total_ns,avg_ns,stddev_ns,min_ns,max_ns", lines[0])
}

func TestWriteCSVStoreStats(t *testing.T) {
	results := &Results{
		StoreStats: map[string]*StoreStat{
			"StoreGetObject": {
				TimingStat:   TimingStat{Count: 5, TotalNs: 1000},
				TotalSize:    500,
				BytesRead:    500,
				BytesWritten: 0,
			},
		},
		TimingEnabled: true,
	}

	var buf bytes.Buffer
	err := results.WriteCSVSection(&buf, SectionStore)
	require.NoError(t, err)

	csv := buf.String()
	assert.Contains(t, csv, "operation,count,bytes_read,bytes_written,total_ns")
	assert.Contains(t, csv, "StoreGetObject,5,500,0,1000")
}
