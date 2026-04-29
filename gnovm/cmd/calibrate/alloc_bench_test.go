package calibrate

import "testing"

// Benchmarks for Go heap allocation cost across sizes.
// Used to calibrate GnoVM's allocation gas model.
//
// Run: go test -bench=BenchmarkAlloc -benchmem -count=5 -timeout=20m
//
// The results feed into the two-tier gas model in alloc.go:
//   - Go's allocator has universal boundaries at 16B (tiny/small) and 32KB (small/large)
//   - Cost grows sublinearly with size (~size^0.78) due to memclr amortization
//   - The model must be conservative: no allocation undercharged >10%

var sink []byte

func benchAlloc(b *testing.B, n int) {
	b.Helper()
	for i := 0; i < b.N; i++ {
		sink = make([]byte, n)
	}
}

// Powers of 2 from 1B to 1GB, plus key size-class boundaries.
func BenchmarkAlloc_1(b *testing.B)          { benchAlloc(b, 1) }
func BenchmarkAlloc_2(b *testing.B)          { benchAlloc(b, 2) }
func BenchmarkAlloc_4(b *testing.B)          { benchAlloc(b, 4) }
func BenchmarkAlloc_8(b *testing.B)          { benchAlloc(b, 8) }
func BenchmarkAlloc_16(b *testing.B)         { benchAlloc(b, 16) }
func BenchmarkAlloc_32(b *testing.B)         { benchAlloc(b, 32) }
func BenchmarkAlloc_64(b *testing.B)         { benchAlloc(b, 64) }
func BenchmarkAlloc_96(b *testing.B)         { benchAlloc(b, 96) }
func BenchmarkAlloc_112(b *testing.B)        { benchAlloc(b, 112) }
func BenchmarkAlloc_128(b *testing.B)        { benchAlloc(b, 128) }
func BenchmarkAlloc_144(b *testing.B)        { benchAlloc(b, 144) }
func BenchmarkAlloc_192(b *testing.B)        { benchAlloc(b, 192) }
func BenchmarkAlloc_256(b *testing.B)        { benchAlloc(b, 256) }
func BenchmarkAlloc_384(b *testing.B)        { benchAlloc(b, 384) }
func BenchmarkAlloc_512(b *testing.B)        { benchAlloc(b, 512) }
func BenchmarkAlloc_768(b *testing.B)        { benchAlloc(b, 768) }
func BenchmarkAlloc_1024(b *testing.B)       { benchAlloc(b, 1024) }
func BenchmarkAlloc_2048(b *testing.B)       { benchAlloc(b, 2048) }
func BenchmarkAlloc_4096(b *testing.B)       { benchAlloc(b, 4096) }
func BenchmarkAlloc_8192(b *testing.B)       { benchAlloc(b, 8192) }
func BenchmarkAlloc_16384(b *testing.B)      { benchAlloc(b, 16384) }
func BenchmarkAlloc_32768(b *testing.B)      { benchAlloc(b, 32768) }
func BenchmarkAlloc_65536(b *testing.B)      { benchAlloc(b, 65536) }
func BenchmarkAlloc_131072(b *testing.B)     { benchAlloc(b, 131072) }
func BenchmarkAlloc_262144(b *testing.B)     { benchAlloc(b, 262144) }
func BenchmarkAlloc_524288(b *testing.B)     { benchAlloc(b, 524288) }
func BenchmarkAlloc_1048576(b *testing.B)    { benchAlloc(b, 1048576) }
func BenchmarkAlloc_2097152(b *testing.B)    { benchAlloc(b, 2097152) }
func BenchmarkAlloc_4194304(b *testing.B)    { benchAlloc(b, 4194304) }
func BenchmarkAlloc_8388608(b *testing.B)    { benchAlloc(b, 8388608) }
func BenchmarkAlloc_16777216(b *testing.B)   { benchAlloc(b, 16777216) }
func BenchmarkAlloc_33554432(b *testing.B)   { benchAlloc(b, 33554432) }
func BenchmarkAlloc_67108864(b *testing.B)   { benchAlloc(b, 67108864) }
func BenchmarkAlloc_134217728(b *testing.B)  { benchAlloc(b, 134217728) }
func BenchmarkAlloc_268435456(b *testing.B)  { benchAlloc(b, 268435456) }
func BenchmarkAlloc_536870912(b *testing.B)  { benchAlloc(b, 536870912) }
func BenchmarkAlloc_1073741824(b *testing.B) { benchAlloc(b, 1073741824) }
