package calibrate

// Native function calibration benchmarks (pure-CPU side, end-to-end via dispatcher).
//
// Each bench drives the GnoVM dispatcher's native-call path:
//   stdlibs.NativeResolver(pkg, name)(m)
//
// This invokes the same closure stored in *FuncValue.nativeBody — the
// generated wrapper from gnovm/stdlibs/generated.go — which performs
// Gno→Go reflective parameter conversion, calls the X_ function, and
// converts return values back. Measurement therefore captures the FULL
// dispatch cost (reflect overhead + X_ work + Go2GnoValue return push).
//
// chargeNativeGas runs in the actual production path before this wrapper.
// We do NOT call it here — the per-function gas formula (Base + Slope*N)
// derived from these benches IS the total dispatcher charge in the new
// runtime model. There's no separate OpCPUNativeDispatch floor.
//
// Run:
//   cd gnovm/cmd/calibrate
//   go test -bench=BenchmarkNative -benchtime=200ms -count=3 -timeout=15m . \
//       > native_bench_output.txt
//   python3 gen_native_table.py native_bench_output.txt

import (
	"crypto/ed25519"
	"crypto/rand"
	"math"
	"reflect"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	tmcrypto "github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// dispatchHarness drives the native dispatcher wrapper and resets the
// value stack between iterations.
type dispatchHarness struct {
	m        *gno.Machine
	wrapper  func(*gno.Machine)
	nReturns int
}

func (h *dispatchHarness) call() {
	h.wrapper(h.m)
	if h.nReturns > 0 {
		_ = h.m.PopValues(h.nReturns)
	}
}

// newDispatchMachine builds a Machine with Alloc + Store + a single Block
// of `nParams` slots. The caller populates Block.Values[i] with TVs built
// via gno.Go2GnoValue. Frames are set up by the caller for natives that
// need them (Context-readers, frame-walkers).
func newDispatchMachine(nParams int) *gno.Machine {
	m := &gno.Machine{
		Alloc: gno.NewAllocator(math.MaxInt64),
		Stage: gno.StageRun,
	}
	m.Blocks = []*gno.Block{{Values: make([]gno.TypedValue, nParams)}}
	return m
}

// setBlockValueFromGo populates Block.Values[idx] with a TV built from a
// reflect.Value of the desired Go type. The wrapper's reflect-based
// Gno2GnoValue conversion will read the result.
func setBlockValueFromGo(m *gno.Machine, idx int, v interface{}) {
	m.Blocks[0].Values[idx] = gno.Go2GnoValue(m.Alloc, m.Store, reflect.ValueOf(v))
}

// resolveWrapper panics if the (pkg, name) isn't registered. Use to fail
// fast in bench setup rather than nil-deref in the hot loop.
func resolveWrapper(b *testing.B, pkg string, name gno.Name) func(*gno.Machine) {
	b.Helper()
	w := stdlibs.NativeResolver(pkg, name)
	if w == nil {
		b.Fatalf("native %s.%s not found", pkg, name)
	}
	return w
}

// ----- crypto/sha256.sum256(data []byte) [32]byte -----

func benchSHA256(b *testing.B, n int) {
	b.Helper()
	data := make([]byte, n)
	rand.Read(data)
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, data)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/sha256", "sum256"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SHA256_Sum256_0(b *testing.B)     { benchSHA256(b, 0) }
func BenchmarkNative_SHA256_Sum256_64(b *testing.B)    { benchSHA256(b, 64) }
func BenchmarkNative_SHA256_Sum256_256(b *testing.B)   { benchSHA256(b, 256) }
func BenchmarkNative_SHA256_Sum256_1024(b *testing.B)  { benchSHA256(b, 1024) }
func BenchmarkNative_SHA256_Sum256_4096(b *testing.B)  { benchSHA256(b, 4096) }
func BenchmarkNative_SHA256_Sum256_16384(b *testing.B) { benchSHA256(b, 16384) }
func BenchmarkNative_SHA256_Sum256_65536(b *testing.B) { benchSHA256(b, 65536) }

// ----- crypto/ed25519.verify(pub, msg, sig []byte) bool -----

func benchEd25519Verify(b *testing.B, msgLen int) {
	b.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	msg := make([]byte, msgLen)
	rand.Read(msg)
	sig := ed25519.Sign(priv, msg)
	m := newDispatchMachine(3)
	setBlockValueFromGo(m, 0, []byte(pub))
	setBlockValueFromGo(m, 1, msg)
	setBlockValueFromGo(m, 2, sig)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/ed25519", "verify"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Ed25519_Verify_64(b *testing.B)    { benchEd25519Verify(b, 64) }
func BenchmarkNative_Ed25519_Verify_256(b *testing.B)   { benchEd25519Verify(b, 256) }
func BenchmarkNative_Ed25519_Verify_1024(b *testing.B)  { benchEd25519Verify(b, 1024) }
func BenchmarkNative_Ed25519_Verify_4096(b *testing.B)  { benchEd25519Verify(b, 4096) }
func BenchmarkNative_Ed25519_Verify_16384(b *testing.B) { benchEd25519Verify(b, 16384) }

// ----- math.Float{32,64}{bits,frombits} -----

func benchMathFlat(b *testing.B, fn gno.Name, paramVal interface{}) {
	b.Helper()
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, paramVal)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "math", fn), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Math_Float32bits(b *testing.B) { benchMathFlat(b, "Float32bits", float32(1.5)) }
func BenchmarkNative_Math_Float32frombits(b *testing.B) {
	benchMathFlat(b, "Float32frombits", uint32(0x3FC00000))
}
func BenchmarkNative_Math_Float64bits(b *testing.B) { benchMathFlat(b, "Float64bits", float64(1.5)) }
func BenchmarkNative_Math_Float64frombits(b *testing.B) {
	benchMathFlat(b, "Float64frombits", uint64(0x3FF8000000000000))
}

// ----- chain.packageAddress(pkgPath string) string -----

func benchChainPackageAddress(b *testing.B, n int) {
	b.Helper()
	pkgPath := strings.Repeat("x", n)
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, pkgPath)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain", "packageAddress"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Chain_PackageAddress_1(b *testing.B)    { benchChainPackageAddress(b, 1) }
func BenchmarkNative_Chain_PackageAddress_10(b *testing.B)   { benchChainPackageAddress(b, 10) }
func BenchmarkNative_Chain_PackageAddress_100(b *testing.B)  { benchChainPackageAddress(b, 100) }
func BenchmarkNative_Chain_PackageAddress_1000(b *testing.B) { benchChainPackageAddress(b, 1000) }

// ----- chain.deriveStorageDepositAddr(pkgPath string) string -----

func benchChainDeriveStorageDepositAddr(b *testing.B, n int) {
	b.Helper()
	pkgPath := strings.Repeat("x", n)
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, pkgPath)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain", "deriveStorageDepositAddr"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Chain_DeriveStorageDepositAddr_1(b *testing.B) {
	benchChainDeriveStorageDepositAddr(b, 1)
}
func BenchmarkNative_Chain_DeriveStorageDepositAddr_10(b *testing.B) {
	benchChainDeriveStorageDepositAddr(b, 10)
}
func BenchmarkNative_Chain_DeriveStorageDepositAddr_100(b *testing.B) {
	benchChainDeriveStorageDepositAddr(b, 100)
}
func BenchmarkNative_Chain_DeriveStorageDepositAddr_1000(b *testing.B) {
	benchChainDeriveStorageDepositAddr(b, 1000)
}

// ----- chain.pubKeyAddress(bech32PubKey string) (addr, errStr string) -----

func BenchmarkNative_Chain_PubKeyAddress(b *testing.B) {
	priv := secp256k1.GenPrivKey()
	pubBech32 := tmcrypto.PubKeyToBech32(priv.PubKey())
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, pubBech32)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain", "pubKeyAddress"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// ----- time.loadFromEmbeddedTZData(name string) (data []byte, found bool) -----

func BenchmarkNative_Time_LoadTZData(b *testing.B) {
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, "America/New_York")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "time", "loadFromEmbeddedTZData"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}
