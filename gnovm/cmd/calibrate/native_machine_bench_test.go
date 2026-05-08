package calibrate

// Native function calibration benchmarks (Machine-harness side).
//
// Companion to native_bench_test.go. Drives the dispatcher wrapper for
// natives that need an ExecContext (Banker, Params, EventLogger) and/or
// a realistic frame stack. Same end-to-end model: full Gno↔Go reflect
// + X_ call cost is captured in the measurement.

import (
	"crypto/rand"
	"fmt"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// ---------------- mock interfaces ----------------

type mockBanker struct {
	coins map[crypto.Bech32Address]std.Coins
}

func newMockBanker() *mockBanker {
	return &mockBanker{coins: map[crypto.Bech32Address]std.Coins{}}
}

// All ops are no-op except GetCoins (which returns the preload). We're
// measuring the X_ wrapper's Go-side cost, not the real Banker's storage
// I/O — that's metered separately by the KVStore's per-byte gas.
func (b *mockBanker) GetCoins(addr crypto.Bech32Address) std.Coins {
	return b.coins[addr]
}
func (b *mockBanker) SendCoins(from, to crypto.Bech32Address, amt std.Coins)           {}
func (b *mockBanker) TotalCoin(denom string) int64                                     { return 0 }
func (b *mockBanker) IssueCoin(addr crypto.Bech32Address, denom string, amount int64)  {}
func (b *mockBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amount int64) {}

type mockParams struct {
	data map[string]any
}

func newMockParams() *mockParams { return &mockParams{data: map[string]any{}} }

func (p *mockParams) SetString(k, v string)           { p.data[k] = v }
func (p *mockParams) SetBool(k string, v bool)        { p.data[k] = v }
func (p *mockParams) SetInt64(k string, v int64)      { p.data[k] = v }
func (p *mockParams) SetUint64(k string, v uint64)    { p.data[k] = v }
func (p *mockParams) SetBytes(k string, v []byte)     { p.data[k] = v }
func (p *mockParams) SetStrings(k string, v []string) { p.data[k] = v }
func (p *mockParams) UpdateStrings(k string, v []string, add bool) {
	if add {
		cur, _ := p.data[k].([]string)
		p.data[k] = append(cur, v...)
	} else {
		p.data[k] = v
	}
}
func (p *mockParams) GetString(k string, ptr *string) bool {
	v, ok := p.data[k].(string)
	if ok {
		*ptr = v
	}
	return ok
}
func (p *mockParams) GetBool(k string, ptr *bool) bool {
	v, ok := p.data[k].(bool)
	if ok {
		*ptr = v
	}
	return ok
}
func (p *mockParams) GetInt64(k string, ptr *int64) bool {
	v, ok := p.data[k].(int64)
	if ok {
		*ptr = v
	}
	return ok
}
func (p *mockParams) GetUint64(k string, ptr *uint64) bool {
	v, ok := p.data[k].(uint64)
	if ok {
		*ptr = v
	}
	return ok
}
func (p *mockParams) GetBytes(k string, ptr *[]byte) bool {
	v, ok := p.data[k].([]byte)
	if ok {
		*ptr = v
	}
	return ok
}
func (p *mockParams) GetStrings(k string, ptr *[]string) bool {
	v, ok := p.data[k].([]string)
	if ok {
		*ptr = v
	}
	return ok
}

// addContextAndFrames extends a dispatch Machine with an ExecContext and a
// stack of pkgPath call frames (last entry is the topmost). Pass empty
// string for a "chain"-edge frame (Frames[0] in real runtime is created
// from MsgCall with PkgPath="").
func addContextAndFrames(m *gno.Machine, pkgPaths ...string) (*mockBanker, *mockParams) {
	for _, p := range pkgPaths {
		f := &gno.FuncValue{}
		fr := gno.Frame{
			Func:        f,
			LastPackage: &gno.PackageValue{PkgPath: p},
		}
		m.Frames = append(m.Frames, fr)
	}
	bk := newMockBanker()
	pm := newMockParams()
	send := std.Coins{}
	spent := std.Coins{}
	m.Context = stdlibs.ExecContext{
		ChainID:         "test-chain",
		ChainDomain:     "gno.land",
		Height:          1,
		Timestamp:       1700000000,
		TimestampNano:   0,
		OriginCaller:    crypto.Bech32Address("g1bench" + strings.Repeat("x", 32)),
		OriginSend:      send,
		OriginSendSpent: &spent,
		Banker:          bk,
		Params:          pm,
		EventLogger:     sdk.NewEventLogger(),
	}
	return bk, pm
}

// ---------------- chain/banker ----------------

// X_bankerSendCoins(m, bt uint8, fromS, toS string, denoms []string, amounts []int64)
func benchBankerSendCoins(b *testing.B, n int) {
	b.Helper()
	denoms := make([]string, n)
	amounts := make([]int64, n)
	for i := 0; i < n; i++ {
		denoms[i] = fmt.Sprintf("d%05d", i)
		amounts[i] = 1
	}
	m := newDispatchMachine(5)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/y")
	setBlockValueFromGo(m, 0, uint8(2)) // btRealmSend
	setBlockValueFromGo(m, 1, "g1from")
	setBlockValueFromGo(m, 2, "g1to")
	setBlockValueFromGo(m, 3, denoms)
	setBlockValueFromGo(m, 4, amounts)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/banker", "bankerSendCoins"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Banker_SendCoins_1(b *testing.B)    { benchBankerSendCoins(b, 1) }
func BenchmarkNative_Banker_SendCoins_10(b *testing.B)   { benchBankerSendCoins(b, 10) }
func BenchmarkNative_Banker_SendCoins_100(b *testing.B)  { benchBankerSendCoins(b, 100) }
func BenchmarkNative_Banker_SendCoins_1000(b *testing.B) { benchBankerSendCoins(b, 1000) }

// X_bankerGetCoins(m, bt uint8, addr string) (denoms []string, amounts []int64)
func benchBankerGetCoins(b *testing.B, n int) {
	b.Helper()
	m := newDispatchMachine(2)
	bk, _ := addContextAndFrames(m, "gno.land/r/x")
	addr := crypto.Bech32Address("g1ownr")
	cs := make(std.Coins, n)
	for i := 0; i < n; i++ {
		cs[i] = std.Coin{Denom: fmt.Sprintf("d%05d", i), Amount: int64(i + 1)}
	}
	bk.coins[addr] = cs
	setBlockValueFromGo(m, 0, uint8(0)) // btReadonly
	setBlockValueFromGo(m, 1, string(addr))
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/banker", "bankerGetCoins"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Banker_GetCoins_1(b *testing.B)    { benchBankerGetCoins(b, 1) }
func BenchmarkNative_Banker_GetCoins_10(b *testing.B)   { benchBankerGetCoins(b, 10) }
func BenchmarkNative_Banker_GetCoins_100(b *testing.B)  { benchBankerGetCoins(b, 100) }
func BenchmarkNative_Banker_GetCoins_1000(b *testing.B) { benchBankerGetCoins(b, 1000) }

func BenchmarkNative_Banker_TotalCoin(b *testing.B) {
	m := newDispatchMachine(2)
	addContextAndFrames(m, "gno.land/r/x")
	setBlockValueFromGo(m, 0, uint8(0))
	setBlockValueFromGo(m, 1, "ugnot")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/banker", "bankerTotalCoin"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Banker_IssueCoin(b *testing.B) {
	m := newDispatchMachine(4)
	addContextAndFrames(m, "gno.land/r/x")
	setBlockValueFromGo(m, 0, uint8(3))
	setBlockValueFromGo(m, 1, "g1issue")
	setBlockValueFromGo(m, 2, "ugnot")
	setBlockValueFromGo(m, 3, int64(1))
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/banker", "bankerIssueCoin"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Banker_RemoveCoin(b *testing.B) {
	m := newDispatchMachine(4)
	addContextAndFrames(m, "gno.land/r/x")
	setBlockValueFromGo(m, 0, uint8(3))
	setBlockValueFromGo(m, 1, "g1issue")
	setBlockValueFromGo(m, 2, "ugnot")
	setBlockValueFromGo(m, 3, int64(1))
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/banker", "bankerRemoveCoin"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Banker_OriginSend(b *testing.B) {
	m := newDispatchMachine(0)
	addContextAndFrames(m, "gno.land/r/x")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/banker", "originSend"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Banker_AssertCallerIsRealm(b *testing.B) {
	m := newDispatchMachine(0)
	addContextAndFrames(m, "gno.land/r/caller", "gno.land/r/callee")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/banker", "assertCallerIsRealm"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// ---------------- chain.emit ----------------
// Bench grid is 2-D: (nAttrs, perElemBytes). The fitter regresses
// cost = base + α·nAttrs + β·totalBytes. Constant-byte benches isolate
// the count slope (α); constant-count benches isolate the byte slope (β).
// emit truncates each value to MaxEventAttrLen=1024, so per-element
// payloads above that cap don't grow the marshal cost — we keep the
// bytes-grid at ≤1024/element so β reflects real marshal work.

func benchChainEmit(b *testing.B, nAttrs, perElemBytes int) {
	b.Helper()
	elem := strings.Repeat("k", perElemBytes)
	attrs := make([]string, nAttrs)
	for i := range attrs {
		attrs[i] = elem
	}
	if len(attrs)%2 != 0 {
		attrs = attrs[:len(attrs)-1]
	}
	m := newDispatchMachine(2)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/y")
	setBlockValueFromGo(m, 0, "T")
	setBlockValueFromGo(m, 1, attrs)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain", "emit"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// Count axis (perElemBytes=1).
func BenchmarkNative_Chain_Emit_2_1(b *testing.B)   { benchChainEmit(b, 2, 1) }
func BenchmarkNative_Chain_Emit_10_1(b *testing.B)  { benchChainEmit(b, 10, 1) }
func BenchmarkNative_Chain_Emit_100_1(b *testing.B) { benchChainEmit(b, 100, 1) }

// 128 = MaxEventPairs * 2 — the new hard cap from emit_event.go.
func BenchmarkNative_Chain_Emit_128_1(b *testing.B) { benchChainEmit(b, 128, 1) }

// Bytes axis (nAttrs=2). 1024 is MaxEventAttrLen; above that emit truncates
// silently so additional bytes don't grow the marshal slope.
func BenchmarkNative_Chain_Emit_2_50(b *testing.B)   { benchChainEmit(b, 2, 50) }
func BenchmarkNative_Chain_Emit_2_500(b *testing.B)  { benchChainEmit(b, 2, 500) }
func BenchmarkNative_Chain_Emit_2_1024(b *testing.B) { benchChainEmit(b, 2, 1024) }

// ---------------- chain/params ----------------
// Per-native cost is X_ wrapper overhead only; per-byte storage cost is
// metered separately by the KVStore via gctx (see params keeper fix).

func benchParamsSetBytes(b *testing.B, n int) {
	b.Helper()
	val := make([]byte, n)
	rand.Read(val)
	m := newDispatchMachine(2)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/x")
	setBlockValueFromGo(m, 0, "k")
	setBlockValueFromGo(m, 1, val)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/params", "SetBytes"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Params_SetBytes_1(b *testing.B)    { benchParamsSetBytes(b, 1) }
func BenchmarkNative_Params_SetBytes_10(b *testing.B)   { benchParamsSetBytes(b, 10) }
func BenchmarkNative_Params_SetBytes_100(b *testing.B)  { benchParamsSetBytes(b, 100) }
func BenchmarkNative_Params_SetBytes_1000(b *testing.B) { benchParamsSetBytes(b, 1000) }

func benchParamsSetString(b *testing.B, n int) {
	b.Helper()
	m := newDispatchMachine(2)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/x")
	setBlockValueFromGo(m, 0, "k")
	setBlockValueFromGo(m, 1, strings.Repeat("x", n))
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/params", "SetString"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Params_SetString_1(b *testing.B)    { benchParamsSetString(b, 1) }
func BenchmarkNative_Params_SetString_10(b *testing.B)   { benchParamsSetString(b, 10) }
func BenchmarkNative_Params_SetString_100(b *testing.B)  { benchParamsSetString(b, 100) }
func BenchmarkNative_Params_SetString_1000(b *testing.B) { benchParamsSetString(b, 1000) }

// 2-D grid: count varies (perElem=1) for α; per-elem varies (count=2)
// for β. Bytes are unbounded (no truncation), so β extends to 50k/element.
func benchParamsSetStrings(b *testing.B, n, perElemBytes int) {
	b.Helper()
	elem := strings.Repeat("x", perElemBytes)
	val := make([]string, n)
	for i := range val {
		val[i] = elem
	}
	m := newDispatchMachine(2)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/x")
	setBlockValueFromGo(m, 0, "k")
	setBlockValueFromGo(m, 1, val)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/params", "SetStrings"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Params_SetStrings_1_1(b *testing.B)     { benchParamsSetStrings(b, 1, 1) }
func BenchmarkNative_Params_SetStrings_10_1(b *testing.B)    { benchParamsSetStrings(b, 10, 1) }
func BenchmarkNative_Params_SetStrings_100_1(b *testing.B)   { benchParamsSetStrings(b, 100, 1) }
func BenchmarkNative_Params_SetStrings_1000_1(b *testing.B)  { benchParamsSetStrings(b, 1000, 1) }
func BenchmarkNative_Params_SetStrings_2_50(b *testing.B)    { benchParamsSetStrings(b, 2, 50) }
func BenchmarkNative_Params_SetStrings_2_500(b *testing.B)   { benchParamsSetStrings(b, 2, 500) }
func BenchmarkNative_Params_SetStrings_2_5000(b *testing.B)  { benchParamsSetStrings(b, 2, 5000) }
func BenchmarkNative_Params_SetStrings_2_50000(b *testing.B) { benchParamsSetStrings(b, 2, 50000) }

func BenchmarkNative_Params_SetBool(b *testing.B) {
	m := newDispatchMachine(2)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/x")
	setBlockValueFromGo(m, 0, "k")
	setBlockValueFromGo(m, 1, true)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/params", "SetBool"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Params_SetInt64(b *testing.B) {
	m := newDispatchMachine(2)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/x")
	setBlockValueFromGo(m, 0, "k")
	setBlockValueFromGo(m, 1, int64(42))
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/params", "SetInt64"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Params_SetUint64(b *testing.B) {
	m := newDispatchMachine(2)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/x")
	setBlockValueFromGo(m, 0, "k")
	setBlockValueFromGo(m, 1, uint64(42))
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/params", "SetUint64"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// ---------------- sys/params ----------------
// assertSysParamsRealm requires:
//   Frames[len-1].LastPackage.PkgPath == "sys/params"
//   Frames[len-2].LastPackage.PkgPath == "gno.land/r/sys/params"

func benchSysParamsSetBytes(b *testing.B, n int) {
	b.Helper()
	val := make([]byte, n)
	rand.Read(val)
	m := newDispatchMachine(4)
	addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	setBlockValueFromGo(m, 3, val)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", "setSysParamBytes"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_SetBytes_1(b *testing.B)    { benchSysParamsSetBytes(b, 1) }
func BenchmarkNative_SysParams_SetBytes_10(b *testing.B)   { benchSysParamsSetBytes(b, 10) }
func BenchmarkNative_SysParams_SetBytes_100(b *testing.B)  { benchSysParamsSetBytes(b, 100) }
func BenchmarkNative_SysParams_SetBytes_1000(b *testing.B) { benchSysParamsSetBytes(b, 1000) }

func benchSysParamsGetBytes(b *testing.B, n int) {
	b.Helper()
	m := newDispatchMachine(3)
	_, pm := addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	val := make([]byte, n)
	rand.Read(val)
	pm.SetBytes("mod:sub:name", val) // pre-seed via mock
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", "getSysParamBytes"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_GetBytes_1(b *testing.B)    { benchSysParamsGetBytes(b, 1) }
func BenchmarkNative_SysParams_GetBytes_10(b *testing.B)   { benchSysParamsGetBytes(b, 10) }
func BenchmarkNative_SysParams_GetBytes_100(b *testing.B)  { benchSysParamsGetBytes(b, 100) }
func BenchmarkNative_SysParams_GetBytes_1000(b *testing.B) { benchSysParamsGetBytes(b, 1000) }

// ---- sys/params: setSysParamString ----

func benchSysParamsSetString(b *testing.B, n int) {
	b.Helper()
	m := newDispatchMachine(4)
	addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	setBlockValueFromGo(m, 3, strings.Repeat("x", n))
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", "setSysParamString"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_SetString_1(b *testing.B)    { benchSysParamsSetString(b, 1) }
func BenchmarkNative_SysParams_SetString_10(b *testing.B)   { benchSysParamsSetString(b, 10) }
func BenchmarkNative_SysParams_SetString_100(b *testing.B)  { benchSysParamsSetString(b, 100) }
func BenchmarkNative_SysParams_SetString_1000(b *testing.B) { benchSysParamsSetString(b, 1000) }

// ---- sys/params: setSysParamStrings ----

func benchSysParamsSetStrings(b *testing.B, n, perElemBytes int) {
	b.Helper()
	elem := strings.Repeat("x", perElemBytes)
	val := make([]string, n)
	for i := range val {
		val[i] = elem
	}
	m := newDispatchMachine(4)
	addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	setBlockValueFromGo(m, 3, val)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", "setSysParamStrings"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_SetStrings_1_1(b *testing.B)    { benchSysParamsSetStrings(b, 1, 1) }
func BenchmarkNative_SysParams_SetStrings_10_1(b *testing.B)   { benchSysParamsSetStrings(b, 10, 1) }
func BenchmarkNative_SysParams_SetStrings_100_1(b *testing.B)  { benchSysParamsSetStrings(b, 100, 1) }
func BenchmarkNative_SysParams_SetStrings_1000_1(b *testing.B) { benchSysParamsSetStrings(b, 1000, 1) }
func BenchmarkNative_SysParams_SetStrings_2_50(b *testing.B)   { benchSysParamsSetStrings(b, 2, 50) }
func BenchmarkNative_SysParams_SetStrings_2_500(b *testing.B)  { benchSysParamsSetStrings(b, 2, 500) }
func BenchmarkNative_SysParams_SetStrings_2_5000(b *testing.B) { benchSysParamsSetStrings(b, 2, 5000) }
func BenchmarkNative_SysParams_SetStrings_2_50000(b *testing.B) {
	benchSysParamsSetStrings(b, 2, 50000)
}

// ---- sys/params: updateSysParamStrings ----

func benchSysParamsUpdateStrings(b *testing.B, n, perElemBytes int) {
	b.Helper()
	elem := strings.Repeat("x", perElemBytes)
	val := make([]string, n)
	for i := range val {
		val[i] = elem
	}
	m := newDispatchMachine(5)
	addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	setBlockValueFromGo(m, 3, val)
	setBlockValueFromGo(m, 4, false) // add=false → replace
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", "updateSysParamStrings"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_UpdateStrings_1_1(b *testing.B) { benchSysParamsUpdateStrings(b, 1, 1) }
func BenchmarkNative_SysParams_UpdateStrings_10_1(b *testing.B) {
	benchSysParamsUpdateStrings(b, 10, 1)
}
func BenchmarkNative_SysParams_UpdateStrings_100_1(b *testing.B) {
	benchSysParamsUpdateStrings(b, 100, 1)
}
func BenchmarkNative_SysParams_UpdateStrings_1000_1(b *testing.B) {
	benchSysParamsUpdateStrings(b, 1000, 1)
}
func BenchmarkNative_SysParams_UpdateStrings_2_50(b *testing.B) {
	benchSysParamsUpdateStrings(b, 2, 50)
}
func BenchmarkNative_SysParams_UpdateStrings_2_500(b *testing.B) {
	benchSysParamsUpdateStrings(b, 2, 500)
}
func BenchmarkNative_SysParams_UpdateStrings_2_5000(b *testing.B) {
	benchSysParamsUpdateStrings(b, 2, 5000)
}
func BenchmarkNative_SysParams_UpdateStrings_2_50000(b *testing.B) {
	benchSysParamsUpdateStrings(b, 2, 50000)
}

// ---- sys/params: flat setters (Bool/Int64/Uint64) ----

func newSysParamsFlatSetBench(b *testing.B, fn gno.Name, val interface{}) *dispatchHarness {
	b.Helper()
	m := newDispatchMachine(4)
	addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	setBlockValueFromGo(m, 3, val)
	return &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", fn), nReturns: 0}
}

func BenchmarkNative_SysParams_SetBool(b *testing.B) {
	h := newSysParamsFlatSetBench(b, "setSysParamBool", true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_SetInt64(b *testing.B) {
	h := newSysParamsFlatSetBench(b, "setSysParamInt64", int64(42))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_SetUint64(b *testing.B) {
	h := newSysParamsFlatSetBench(b, "setSysParamUint64", uint64(42))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// ---- sys/params: flat getters (Bool/Int64/Uint64) ----

func newSysParamsFlatGetBench(b *testing.B, fn gno.Name, seed func(*mockParams)) *dispatchHarness {
	b.Helper()
	m := newDispatchMachine(3)
	_, pm := addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	if seed != nil {
		seed(pm)
	}
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	return &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", fn), nReturns: 2}
}

func BenchmarkNative_SysParams_GetBool(b *testing.B) {
	h := newSysParamsFlatGetBench(b, "getSysParamBool", func(p *mockParams) { p.SetBool("mod:sub:name", true) })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_GetInt64(b *testing.B) {
	h := newSysParamsFlatGetBench(b, "getSysParamInt64", func(p *mockParams) { p.SetInt64("mod:sub:name", 42) })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_GetUint64(b *testing.B) {
	h := newSysParamsFlatGetBench(b, "getSysParamUint64", func(p *mockParams) { p.SetUint64("mod:sub:name", 42) })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// ---- sys/params: getSysParamString ---- (post-call slope on returned string len)

func benchSysParamsGetString(b *testing.B, n int) {
	b.Helper()
	m := newDispatchMachine(3)
	_, pm := addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	pm.SetString("mod:sub:name", strings.Repeat("x", n))
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", "getSysParamString"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_GetString_1(b *testing.B)    { benchSysParamsGetString(b, 1) }
func BenchmarkNative_SysParams_GetString_10(b *testing.B)   { benchSysParamsGetString(b, 10) }
func BenchmarkNative_SysParams_GetString_100(b *testing.B)  { benchSysParamsGetString(b, 100) }
func BenchmarkNative_SysParams_GetString_1000(b *testing.B) { benchSysParamsGetString(b, 1000) }

// ---- sys/params: getSysParamStrings ---- (post-call slope on returned []string len)

func benchSysParamsGetStrings(b *testing.B, n, perElemBytes int) {
	b.Helper()
	elem := strings.Repeat("x", perElemBytes)
	val := make([]string, n)
	for i := range val {
		val[i] = elem
	}
	m := newDispatchMachine(3)
	_, pm := addContextAndFrames(m, "gno.land/r/sys/params", "sys/params")
	pm.SetStrings("mod:sub:name", val)
	setBlockValueFromGo(m, 0, "mod")
	setBlockValueFromGo(m, 1, "sub")
	setBlockValueFromGo(m, 2, "name")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "sys/params", "getSysParamStrings"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_SysParams_GetStrings_1_1(b *testing.B)    { benchSysParamsGetStrings(b, 1, 1) }
func BenchmarkNative_SysParams_GetStrings_10_1(b *testing.B)   { benchSysParamsGetStrings(b, 10, 1) }
func BenchmarkNative_SysParams_GetStrings_100_1(b *testing.B)  { benchSysParamsGetStrings(b, 100, 1) }
func BenchmarkNative_SysParams_GetStrings_1000_1(b *testing.B) { benchSysParamsGetStrings(b, 1000, 1) }
func BenchmarkNative_SysParams_GetStrings_2_50(b *testing.B)   { benchSysParamsGetStrings(b, 2, 50) }
func BenchmarkNative_SysParams_GetStrings_2_500(b *testing.B)  { benchSysParamsGetStrings(b, 2, 500) }
func BenchmarkNative_SysParams_GetStrings_2_5000(b *testing.B) { benchSysParamsGetStrings(b, 2, 5000) }
func BenchmarkNative_SysParams_GetStrings_2_50000(b *testing.B) {
	benchSysParamsGetStrings(b, 2, 50000)
}

// ---- chain/params: UpdateParamStrings ----

func benchParamsUpdateStrings(b *testing.B, n, perElemBytes int) {
	b.Helper()
	elem := strings.Repeat("x", perElemBytes)
	val := make([]string, n)
	for i := range val {
		val[i] = elem
	}
	m := newDispatchMachine(3)
	addContextAndFrames(m, "gno.land/r/x", "gno.land/r/x")
	setBlockValueFromGo(m, 0, "k")
	setBlockValueFromGo(m, 1, val)
	setBlockValueFromGo(m, 2, false)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/params", "UpdateParamStrings"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Params_UpdateStrings_1_1(b *testing.B)    { benchParamsUpdateStrings(b, 1, 1) }
func BenchmarkNative_Params_UpdateStrings_10_1(b *testing.B)   { benchParamsUpdateStrings(b, 10, 1) }
func BenchmarkNative_Params_UpdateStrings_100_1(b *testing.B)  { benchParamsUpdateStrings(b, 100, 1) }
func BenchmarkNative_Params_UpdateStrings_1000_1(b *testing.B) { benchParamsUpdateStrings(b, 1000, 1) }
func BenchmarkNative_Params_UpdateStrings_2_50(b *testing.B)   { benchParamsUpdateStrings(b, 2, 50) }
func BenchmarkNative_Params_UpdateStrings_2_500(b *testing.B)  { benchParamsUpdateStrings(b, 2, 500) }
func BenchmarkNative_Params_UpdateStrings_2_5000(b *testing.B) { benchParamsUpdateStrings(b, 2, 5000) }
func BenchmarkNative_Params_UpdateStrings_2_50000(b *testing.B) {
	benchParamsUpdateStrings(b, 2, 50000)
}

// ---------------- chain/runtime ----------------

func newRuntimeBench(b *testing.B, fn gno.Name, nReturns int) *dispatchHarness {
	b.Helper()
	m := newDispatchMachine(0)
	addContextAndFrames(m, "gno.land/r/x")
	return &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/runtime", fn), nReturns: nReturns}
}

func BenchmarkNative_Runtime_ChainID(b *testing.B) {
	h := newRuntimeBench(b, "ChainID", 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Runtime_ChainDomain(b *testing.B) {
	h := newRuntimeBench(b, "ChainDomain", 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Runtime_ChainHeight(b *testing.B) {
	h := newRuntimeBench(b, "ChainHeight", 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Runtime_OriginCaller(b *testing.B) {
	h := newRuntimeBench(b, "originCaller", 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Runtime_GetSessionInfo(b *testing.B) {
	h := newRuntimeBench(b, "getSessionInfo", 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// AssertOriginCall: requires Frames[0].LastPackage.PkgPath == "" and
// NumCallFrames() <= 2.
func BenchmarkNative_Runtime_AssertOriginCall(b *testing.B) {
	m := newDispatchMachine(0)
	addContextAndFrames(m, "", "gno.land/r/x")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/runtime", "AssertOriginCall"), nReturns: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// getRealm(height): walks frames looking for the (height+1)-th WithCross.
// We set NO WithCross, height=0 — loop walks the full stack then falls
// through to the StageRun switch (crosses=0, height=0 → derives address
// from Frames[0].LastPackage.PkgPath). Cost = O(depth) loop + 1 sha256.
func benchRuntimeGetRealm(b *testing.B, depth int) {
	b.Helper()
	pkgs := make([]string, depth)
	for i := range pkgs {
		pkgs[i] = fmt.Sprintf("gno.land/r/p%d", i)
	}
	m := newDispatchMachine(1)
	addContextAndFrames(m, pkgs...)
	setBlockValueFromGo(m, 0, 0) // height
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/runtime", "getRealm"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Runtime_GetRealm_1(b *testing.B)    { benchRuntimeGetRealm(b, 1) }
func BenchmarkNative_Runtime_GetRealm_10(b *testing.B)   { benchRuntimeGetRealm(b, 10) }
func BenchmarkNative_Runtime_GetRealm_100(b *testing.B)  { benchRuntimeGetRealm(b, 100) }
func BenchmarkNative_Runtime_GetRealm_1000(b *testing.B) { benchRuntimeGetRealm(b, 1000) }

// ---------------- time.now ----------------

func BenchmarkNative_Time_Now(b *testing.B) {
	m := newDispatchMachine(0)
	addContextAndFrames(m, "gno.land/r/x")
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "time", "now"), nReturns: 3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}
