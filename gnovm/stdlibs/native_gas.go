package stdlibs

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// Per-native gas charging. Calibrated values come from
// gnovm/cmd/calibrate/native_bench_test.go (and its machine-harness
// companion); the python fitter at gnovm/cmd/calibrate/gen_native_table.py
// emits the literal table below.
//
// Registration is at init time into a package global in gnolang
// (RegisterNativeGas). Mirrors how OpCPU* constants live in gnolang —
// no per-Machine resolver plumbing.
//
// Calibration is end-to-end through doOpCallNativeBody (Gno↔Go reflect +
// X_ work + return push); no separate dispatch overhead constant.
//
// Two-slope (Slope2/PostSlope2) is supported by the runtime for natives
// whose cost depends on both element count and per-element bytes (e.g.
// hypothetical natives that hash inside the dispatcher). For all natives
// shipped today the empirical per-byte CPU cost is negligible — the
// per-byte work happens inside the metered KVStore (gctx) — so every row
// below uses a single slope.

// Re-export the gnolang SizeKind constants for readable table literals.
const (
	SizeFlat            = gno.SizeFlat
	SizeLenBytes        = gno.SizeLenBytes
	SizeLenString       = gno.SizeLenString
	SizeLenSlice        = gno.SizeLenSlice
	SizeNumCallFrames   = gno.SizeNumCallFrames
	SizeReturnLen       = gno.SizeReturnLen
	SizeSliceTotalBytes = gno.SizeSliceTotalBytes
)

// nativeGasEntry is the on-disk shape of a row, copied into a
// gno.NativeGasInfo at init time.
type nativeGasEntry struct {
	Pkg, Fn   string
	Base      int64
	Slope     int64
	SlopeIdx  int8
	SlopeKind gno.NativeGasSize

	// Optional second pre-call slope for natives whose cost depends on
	// two independent dimensions (e.g. count and total inner bytes).
	Slope2     int64
	Slope2Idx  int8
	Slope2Kind gno.NativeGasSize

	// Post-call charge (zero = none). PostSlopeIdx is the stack offset
	// from top (1 = last-pushed return); PostSlopeKind must be a
	// SizeReturn* kind (or SizeSliceTotalBytes for slice-of-string).
	PostBase      int64
	PostSlope     int64
	PostSlopeIdx  int8
	PostSlopeKind gno.NativeGasSize

	// Optional second post-call slope, summed independently.
	PostSlope2     int64
	PostSlope2Idx  int8
	PostSlope2Kind gno.NativeGasSize
}

// Calibrated on Apple M2 ARM64 (NOT the reference Xeon 8168 — re-run
// gnovm/cmd/calibrate before any consensus-relevant deployment). 1 gas
// = 1 ns. Slope is ns per 1024 units of N. R² > 0.99 for all linear fits.
//
// Values come from gen_native_table.py over native_bench_output.txt
// captured in a clean-environment bench run. The 2D bench-grid extension
// added later (slice natives benched at multiple per-element byte sizes)
// confirmed the per-byte CPU slope is below noise for every native ship-
// ping today, so the table stays single-slope; the schema fields support
// future natives that genuinely scale on both dimensions.
//
// 46 entries — exhaustive coverage of gnovm/stdlibs/generated.go.
var calibratedNativeGas = []nativeGasEntry{
	{Pkg: "crypto/sha256", Fn: "sum256", Base: 206, Slope: 8865, SlopeIdx: 0, SlopeKind: SizeLenBytes}, // fit base=206.0ns slope=8.6575ns/N (=8865/1024) R²=1.000
	{Pkg: "crypto/ed25519", Fn: "verify", Base: 56407, Slope: 9246, SlopeIdx: 1, SlopeKind: SizeLenBytes}, // fit base=56407.0ns slope=9.0296ns/N (=9246/1024) R²=0.993
	{Pkg: "chain", Fn: "packageAddress", Base: 547, Slope: 15019, SlopeIdx: 0, SlopeKind: SizeLenString}, // fit base=547.5ns slope=14.6672ns/N (=15019/1024) R²=0.999
	{Pkg: "chain", Fn: "deriveStorageDepositAddr", Base: 515, Slope: 453, SlopeIdx: 0, SlopeKind: SizeLenString}, // fit base=515.2ns slope=0.4422ns/N (=453/1024) R²=1.000
	{Pkg: "chain", Fn: "pubKeyAddress", Base: 2571, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 2571.0ns
	{Pkg: "time", Fn: "loadFromEmbeddedTZData", Base: 15507, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 15507.0ns
	{Pkg: "math", Fn: "Float32bits", Base: 38, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 38.0ns
	{Pkg: "math", Fn: "Float32frombits", Base: 33, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 33.0ns
	{Pkg: "math", Fn: "Float64bits", Base: 29, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 29.2ns
	{Pkg: "math", Fn: "Float64frombits", Base: 30, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 30.1ns
	{Pkg: "chain/banker", Fn: "bankerSendCoins", Base: 314, Slope: 34518, SlopeIdx: 3, SlopeKind: SizeLenSlice}, // fit base=314.5ns slope=33.7086ns/N (=34518/1024) R²=0.999
	{Pkg: "chain/banker", Fn: "bankerGetCoins", Base: 340, SlopeIdx: -1, SlopeKind: SizeFlat, PostSlope: 36283, PostSlopeIdx: 2, PostSlopeKind: SizeReturnLen}, // post-call: base=339.5ns + 35.4328ns/N (=36283/1024) R²=0.998
	{Pkg: "chain/banker", Fn: "bankerTotalCoin", Base: 87, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 87.0ns
	{Pkg: "chain/banker", Fn: "bankerIssueCoin", Base: 136, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 136.4ns
	{Pkg: "chain/banker", Fn: "bankerRemoveCoin", Base: 137, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 137.3ns
	{Pkg: "chain/banker", Fn: "originSend", Base: 197, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 196.8ns
	{Pkg: "chain/banker", Fn: "assertCallerIsRealm", Base: 467, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 467.0ns
	{Pkg: "chain", Fn: "emit", Base: 286, Slope: 31192, SlopeIdx: 1, SlopeKind: SizeLenSlice}, // fit base=286.3ns slope=30.4606ns/N (=31192/1024) R²=0.998
	{Pkg: "chain/params", Fn: "SetBytes", Base: 1192, Slope: 8551, SlopeIdx: 1, SlopeKind: SizeLenBytes}, // fit base=1192.0ns slope=8.3504ns/N (=8551/1024) R²=1.000
	{Pkg: "chain/params", Fn: "SetString", Base: 1135, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 1135.0ns
	{Pkg: "chain/params", Fn: "SetStrings", Base: 1205, Slope: 23187, SlopeIdx: 1, SlopeKind: SizeLenSlice}, // fit base=1205.0ns slope=22.6433ns/N (=23187/1024) R²=1.000
	{Pkg: "chain/params", Fn: "SetBool", Base: 1100, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 1100.0ns
	{Pkg: "chain/params", Fn: "SetInt64", Base: 1108, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 1108.0ns
	{Pkg: "chain/params", Fn: "SetUint64", Base: 1106, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 1106.0ns
	{Pkg: "sys/params", Fn: "setSysParamBytes", Base: 290, Slope: 8536, SlopeIdx: 3, SlopeKind: SizeLenBytes}, // fit base=290.3ns slope=8.3360ns/N (=8536/1024) R²=1.000 — val is param 3 of 4: (module, submodule, name, val)
	{Pkg: "sys/params", Fn: "getSysParamBytes", Base: 340, SlopeIdx: -1, SlopeKind: SizeFlat, PostSlope: 10260, PostSlopeIdx: 2, PostSlopeKind: SizeReturnLen}, // post-call: base=339.9ns + 10.0192ns/N (=10260/1024) R²=0.999
	{Pkg: "sys/params", Fn: "setSysParamString", Base: 239, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 239.4ns
	{Pkg: "sys/params", Fn: "setSysParamStrings", Base: 304, Slope: 22949, SlopeIdx: 3, SlopeKind: SizeLenSlice}, // fit base=304.0ns slope=22.4114ns/N (=22949/1024) R²=0.999 — val is param 3 of 4
	{Pkg: "sys/params", Fn: "updateSysParamStrings", Base: 328, Slope: 22840, SlopeIdx: 3, SlopeKind: SizeLenSlice}, // fit base=328.1ns slope=22.3049ns/N (=22840/1024) R²=1.000 — val is param 3 of 5: (module, submodule, name, val, add)
	{Pkg: "sys/params", Fn: "setSysParamBool", Base: 198, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 197.5ns
	{Pkg: "sys/params", Fn: "setSysParamInt64", Base: 209, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 209.4ns
	{Pkg: "sys/params", Fn: "setSysParamUint64", Base: 210, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 209.9ns
	{Pkg: "sys/params", Fn: "getSysParamBool", Base: 202, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 202.4ns
	{Pkg: "sys/params", Fn: "getSysParamInt64", Base: 206, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 205.9ns
	{Pkg: "sys/params", Fn: "getSysParamUint64", Base: 208, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 208.4ns
	{Pkg: "sys/params", Fn: "getSysParamString", Base: 224, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 224.2ns
	{Pkg: "sys/params", Fn: "getSysParamStrings", Base: 364, SlopeIdx: -1, SlopeKind: SizeFlat, PostSlope: 23255, PostSlopeIdx: 2, PostSlopeKind: SizeReturnLen}, // post-call: base=363.9ns + 22.7099ns/N (=23255/1024) R²=1.000
	{Pkg: "chain/params", Fn: "UpdateParamStrings", Base: 1253, Slope: 22784, SlopeIdx: 1, SlopeKind: SizeLenSlice}, // fit base=1253.0ns slope=22.2500ns/N (=22784/1024) R²=1.000
	{Pkg: "chain/runtime", Fn: "ChainID", Base: 44, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 43.8ns
	{Pkg: "chain/runtime", Fn: "ChainDomain", Base: 44, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 43.8ns
	{Pkg: "chain/runtime", Fn: "ChainHeight", Base: 30, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 30.0ns
	{Pkg: "chain/runtime", Fn: "originCaller", Base: 44, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 44.5ns
	{Pkg: "chain/runtime", Fn: "getSessionInfo", Base: 144, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 144.1ns
	{Pkg: "chain/runtime", Fn: "AssertOriginCall", Base: 5, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 5.0ns
	{Pkg: "chain/runtime", Fn: "getRealm", Base: 986, Slope: 1311, SlopeIdx: -1, SlopeKind: SizeNumCallFrames}, // fit base=986.4ns slope=1.2807ns/N (=1311/1024) R²=0.995
	{Pkg: "time", Fn: "now", Base: 47, SlopeIdx: -1, SlopeKind: SizeFlat}, // flat, median 47.1ns
}

func init() {
	for _, e := range calibratedNativeGas {
		gno.RegisterNativeGas(e.Pkg, gno.Name(e.Fn), &gno.NativeGasInfo{
			Base:           e.Base,
			Slope:          e.Slope,
			SlopeIdx:       e.SlopeIdx,
			SlopeKind:      e.SlopeKind,
			Slope2:         e.Slope2,
			Slope2Idx:      e.Slope2Idx,
			Slope2Kind:     e.Slope2Kind,
			PostBase:       e.PostBase,
			PostSlope:      e.PostSlope,
			PostSlopeIdx:   e.PostSlopeIdx,
			PostSlopeKind:  e.PostSlopeKind,
			PostSlope2:     e.PostSlope2,
			PostSlope2Idx:  e.PostSlope2Idx,
			PostSlope2Kind: e.PostSlope2Kind,
		})
	}
}
