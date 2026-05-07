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
// = 1 ns. Slope is ns per 1024 units of N. R² > 0.93 for all linear fits.
//
// Values come from gen_native_table.py over native_bench_output.txt
// (current contents). Production matches the regenerated
// native_gas_table.go.txt exactly — re-running the fitter on this same
// input reproduces this table verbatim. The 2D bench-grid extension
// (slice natives benched at multiple per-element byte sizes) confirmed
// the per-byte CPU slope is below noise for every native shipping
// today, so the table stays single-slope; the schema fields support
// future natives that genuinely scale on both dimensions.
//
// 46 entries — exhaustive coverage of gnovm/stdlibs/generated.go.
var calibratedNativeGas = []nativeGasEntry{
	{Pkg: "crypto/sha256", Fn: "sum256", Base: 226, Slope: 8906, SlopeIdx: 0, SlopeKind: SizeLenBytes},                                                           // fit base=226.3ns slope=8.6969ns/N (=8906/1024) R²=1.000
	{Pkg: "crypto/ed25519", Fn: "verify", Base: 56534, Slope: 8975, SlopeIdx: 1, SlopeKind: SizeLenBytes},                                                        // fit base=56534.0ns slope=8.7645ns/N (=8975/1024) R²=0.991
	{Pkg: "chain", Fn: "packageAddress", Base: 552, Slope: 15201, SlopeIdx: 0, SlopeKind: SizeLenString},                                                         // fit base=552.1ns slope=14.8448ns/N (=15201/1024) R²=0.998
	{Pkg: "chain", Fn: "deriveStorageDepositAddr", Base: 541, Slope: 471, SlopeIdx: 0, SlopeKind: SizeLenString},                                                 // fit base=540.9ns slope=0.4602ns/N (=471/1024) R²=0.994
	{Pkg: "chain", Fn: "pubKeyAddress", Base: 2631, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                           // flat, median 2631.0ns
	{Pkg: "time", Fn: "loadFromEmbeddedTZData", Base: 16068, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                  // flat, median 16068.0ns
	{Pkg: "math", Fn: "Float32bits", Base: 32, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                                // flat, median 32.5ns
	{Pkg: "math", Fn: "Float32frombits", Base: 32, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                            // flat, median 32.4ns
	{Pkg: "math", Fn: "Float64bits", Base: 29, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                                // flat, median 28.7ns
	{Pkg: "math", Fn: "Float64frombits", Base: 29, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                            // flat, median 28.8ns
	{Pkg: "chain/banker", Fn: "bankerSendCoins", Base: 322, Slope: 35318, SlopeIdx: 3, SlopeKind: SizeLenSlice},                                                  // fit base=321.9ns slope=34.4898ns/N (=35318/1024) R²=0.999
	{Pkg: "chain/banker", Fn: "bankerGetCoins", Base: 349, SlopeIdx: -1, SlopeKind: SizeFlat, PostSlope: 36206, PostSlopeIdx: 2, PostSlopeKind: SizeReturnLen},   // post-call: base=349.1ns + 35.3578ns/N (=36206/1024) R²=0.998
	{Pkg: "chain/banker", Fn: "bankerTotalCoin", Base: 89, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                    // flat, median 88.6ns
	{Pkg: "chain/banker", Fn: "bankerIssueCoin", Base: 141, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                   // flat, median 140.6ns
	{Pkg: "chain/banker", Fn: "bankerRemoveCoin", Base: 196, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                  // flat, median 195.9ns
	{Pkg: "chain/banker", Fn: "originSend", Base: 280, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                        // flat, median 280.4ns
	{Pkg: "chain/banker", Fn: "assertCallerIsRealm", Base: 701, SlopeIdx: -1, SlopeKind: SizeFlat},                                                               // flat, median 700.8ns
	{Pkg: "chain/params", Fn: "SetBytes", Base: 1912, Slope: 13213, SlopeIdx: 1, SlopeKind: SizeLenBytes},                                                        // fit base=1912.0ns slope=12.9035ns/N (=13213/1024) R²=1.000
	{Pkg: "chain/params", Fn: "SetString", Base: 1772, Slope: 135, SlopeIdx: 1, SlopeKind: SizeLenString},                                                        // fit base=1772.3ns slope=0.1323ns/N (=135/1024) R²=0.933
	{Pkg: "chain/params", Fn: "SetBool", Base: 1643, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                          // flat, median 1643.0ns
	{Pkg: "chain/params", Fn: "SetInt64", Base: 1201, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                         // flat, median 1201.0ns
	{Pkg: "chain/params", Fn: "SetUint64", Base: 1219, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                        // flat, median 1219.0ns
	{Pkg: "sys/params", Fn: "setSysParamBytes", Base: 323, Slope: 9703, SlopeIdx: 3, SlopeKind: SizeLenBytes},                                                    // fit base=323.3ns slope=9.4757ns/N (=9703/1024) R²=0.995
	{Pkg: "sys/params", Fn: "getSysParamBytes", Base: 416, SlopeIdx: -1, SlopeKind: SizeFlat, PostSlope: 10584, PostSlopeIdx: 2, PostSlopeKind: SizeReturnLen},   // post-call: base=415.7ns + 10.3357ns/N (=10584/1024) R²=1.000
	{Pkg: "sys/params", Fn: "setSysParamString", Base: 269, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                   // flat, median 269.1ns
	{Pkg: "sys/params", Fn: "setSysParamBool", Base: 217, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                     // flat, median 217.4ns
	{Pkg: "sys/params", Fn: "setSysParamInt64", Base: 228, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                    // flat, median 227.8ns
	{Pkg: "sys/params", Fn: "setSysParamUint64", Base: 299, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                   // flat, median 298.9ns
	{Pkg: "sys/params", Fn: "getSysParamBool", Base: 236, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                     // flat, median 236.5ns
	{Pkg: "sys/params", Fn: "getSysParamInt64", Base: 323, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                    // flat, median 322.9ns
	{Pkg: "sys/params", Fn: "getSysParamUint64", Base: 309, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                   // flat, median 308.7ns
	{Pkg: "sys/params", Fn: "getSysParamString", Base: 363, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                   // flat, median 362.8ns
	{Pkg: "chain/runtime", Fn: "ChainID", Base: 45, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                           // flat, median 44.8ns
	{Pkg: "chain/runtime", Fn: "ChainDomain", Base: 45, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                       // flat, median 44.5ns
	{Pkg: "chain/runtime", Fn: "ChainHeight", Base: 30, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                       // flat, median 30.2ns
	{Pkg: "chain/runtime", Fn: "originCaller", Base: 45, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                      // flat, median 44.9ns
	{Pkg: "chain/runtime", Fn: "getSessionInfo", Base: 148, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                   // flat, median 148.4ns
	{Pkg: "chain/runtime", Fn: "AssertOriginCall", Base: 5, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                   // flat, median 5.0ns
	{Pkg: "chain/runtime", Fn: "getRealm", Base: 1003, Slope: 1319, SlopeIdx: -1, SlopeKind: SizeNumCallFrames},                                                  // fit base=1003.0ns slope=1.2880ns/N (=1319/1024) R²=0.995
	{Pkg: "time", Fn: "now", Base: 47, SlopeIdx: -1, SlopeKind: SizeFlat},                                                                                        // flat, median 46.9ns
	{Pkg: "chain", Fn: "emit", Base: 362, Slope: 40218, SlopeIdx: 1, SlopeKind: SizeLenSlice},                                                                    // fit base=361.9ns slope=39.2750ns/N (=40218/1024) R²=0.955
	{Pkg: "chain/params", Fn: "SetStrings", Base: 1601, Slope: 39842, SlopeIdx: 1, SlopeKind: SizeLenSlice},                                                      // fit base=1601.1ns slope=38.9082ns/N (=39842/1024) R²=0.993
	{Pkg: "chain/params", Fn: "UpdateParamStrings", Base: 1298, Slope: 24077, SlopeIdx: 1, SlopeKind: SizeLenSlice},                                              // fit base=1298.0ns slope=23.5122ns/N (=24077/1024) R²=1.000
	{Pkg: "sys/params", Fn: "setSysParamStrings", Base: 341, Slope: 27034, SlopeIdx: 3, SlopeKind: SizeLenSlice},                                                 // fit base=341.0ns slope=26.4006ns/N (=27034/1024) R²=0.997
	{Pkg: "sys/params", Fn: "updateSysParamStrings", Base: 413, Slope: 26861, SlopeIdx: 3, SlopeKind: SizeLenSlice},                                              // fit base=413.4ns slope=26.2318ns/N (=26861/1024) R²=0.998
	{Pkg: "sys/params", Fn: "getSysParamStrings", Base: 349, SlopeIdx: -1, SlopeKind: SizeFlat, PostSlope: 23215, PostSlopeIdx: 2, PostSlopeKind: SizeReturnLen}, // post-call: base=348.9ns + 22.6713ns/N (=23215/1024) R²=0.999
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
