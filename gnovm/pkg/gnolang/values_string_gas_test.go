package gnolang

import (
	"bytes"
	"io"
	"testing"
)

// Calibration benchmarks for the per-flush "stream output" gas charge in
// meteredWriter.Flush (see streamOutputGas in values_string_stream.go).
//
// Reference-hardware convention: 1 gas = 1 ns on the DigitalOcean Dedicated
// 2-core box (Intel Xeon 8168 @ 2.70GHz, Go 1.24/linux/amd64); see
// gnovm/cmd/calibrate. This machine is usually NOT the reference box, so the
// constant is derived with the rule-of-three against a known-calibrated
// anchor — a 1 KiB heap allocation, whose reference cost is
// allocGasTable[10] = 241 gas:
//
//	gasPerOutputByte_ref = produceNsPerByte_local * (241 / anchor1KBNs_local)
//
// Procedure:
//   go test -run X -bench 'StreamOutput' -benchmem ./gnovm/pkg/gnolang/
// then for each Produce benchmark compute ns/byte = (ns/op) / (bytes/op),
// scale by (241 / anchorNsPerOp), and take a representative (rounded-up)
// integer for streamOutputGasPerByte.

var gasBenchSink []byte

// BenchmarkStreamOutputAnchor1KBAlloc is the rule-of-three anchor: one 1 KiB
// heap allocation. Its reference gas cost is allocGasTable[10] = 241.
func BenchmarkStreamOutputAnchor1KBAlloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gasBenchSink = make([]byte, 1024)
	}
	_ = gasBenchSink
}

// benchStreamOutputProduce times the end-to-end cost of formatting AND
// flushing a []int of n elements into a discarding metered writer with no gas
// meter (we measure the work, not the accounting). It reports bytes/op so that
// ns/byte = (ns/op) / (bytes/op) — the per-output-byte production cost the
// per-flush charge is meant to price.
func benchStreamOutputProduce(b *testing.B, n int) {
	b.Helper()
	elems := make([]TypedValue, n)
	for i := range elems {
		elems[i] = typedInt(i)
	}
	tv := typedSlice(IntType, elems...)
	nbytes := len(tv.String())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mw := newUnmeteredWriter(io.Discard)
		tv.Fprint(mw, nil)
		mw.Flush()
		mw.Release()
	}
	b.StopTimer()
	b.ReportMetric(float64(nbytes), "bytes/op")
}

func BenchmarkStreamOutputProduce_Int100(b *testing.B)   { benchStreamOutputProduce(b, 100) }
func BenchmarkStreamOutputProduce_Int1000(b *testing.B)  { benchStreamOutputProduce(b, 1000) }
func BenchmarkStreamOutputProduce_Int10000(b *testing.B) { benchStreamOutputProduce(b, 10000) }

// BenchmarkStreamOutputFlushOnly isolates the raw flush memcpy (a full buffer
// drained to io.Discard) — the lower bound on per-byte cost, excluding
// formatting. ns/byte = (ns/op) / meteredWriterBufSize. Reported for context;
// the production benchmarks above are what set the constant.
func BenchmarkStreamOutputFlushOnly(b *testing.B) {
	full := bytes.Repeat([]byte("x"), meteredWriterBufSize)
	mw := newUnmeteredWriter(io.Discard)
	defer mw.Release()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mw.WriteBytes(full)
		mw.Flush()
	}
}
