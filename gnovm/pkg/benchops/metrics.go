package benchops

import (
	"math"
	"strconv"
	"time"
)

// ---- Unified stat types (used for both collection and export)

// TimingStat contains common timing statistics.
// Used for both runtime collection and JSON export.
type TimingStat struct {
	Count   int64 `json:"count"`
	TotalNs int64 `json:"total_ns,omitempty"`
	MinNs   int64 `json:"min_ns,omitempty"`
	MaxNs   int64 `json:"max_ns,omitempty"`
	SumSqNs int64 `json:"sumsq_ns,omitempty"` // sum of squared durations for stddev/merge
}

// Record adds a sample to the timing statistics.
// The duration may be zero (for count-only recording when timing is disabled).
func (t *TimingStat) Record(dur time.Duration) {
	t.Count++
	if dur <= 0 {
		return
	}
	ns := dur.Nanoseconds()
	t.TotalNs += ns
	if t.MinNs == 0 || ns < t.MinNs {
		t.MinNs = ns
	}
	if ns > t.MaxNs {
		t.MaxNs = ns
	}
	t.SumSqNs += ns * ns
}

// AvgNs returns the average duration in nanoseconds.
func (t *TimingStat) AvgNs() int64 {
	if t.Count == 0 {
		return 0
	}
	return t.TotalNs / t.Count
}

// StdDevNs returns the standard deviation in nanoseconds.
// Uses the computational formula: stddev = sqrt(E[X^2] - E[X]^2)
// Returns 0 if count < 2 (std dev undefined for single sample).
func (t *TimingStat) StdDevNs() int64 {
	if t.Count < 2 {
		return 0
	}
	meanNs := float64(t.TotalNs) / float64(t.Count)
	variance := float64(t.SumSqNs)/float64(t.Count) - meanNs*meanNs
	if variance < 0 {
		// Numerical instability protection
		variance = 0
	}
	return int64(math.Sqrt(variance))
}

// Merge combines another TimingStat into this one.
func (t *TimingStat) Merge(other *TimingStat) {
	t.Count += other.Count
	t.TotalNs += other.TotalNs
	t.SumSqNs += other.SumSqNs
	if other.MinNs > 0 && (t.MinNs == 0 || other.MinNs < t.MinNs) {
		t.MinNs = other.MinNs
	}
	if other.MaxNs > t.MaxNs {
		t.MaxNs = other.MaxNs
	}
}

// OpStat tracks statistics for a single opcode.
type OpStat struct {
	TimingStat
	Gas int64 `json:"gas"`
}

// Record adds an opcode sample with gas and optional duration.
func (s *OpStat) Record(gas int64, dur time.Duration) {
	s.Gas += gas
	s.TimingStat.Record(dur)
}

// StoreStat tracks statistics for a single store operation.
type StoreStat struct {
	TimingStat
	TotalSize int64 `json:"total_size"`
}

// Record adds a store sample with size and optional duration.
func (s *StoreStat) Record(size int, dur time.Duration) {
	s.TotalSize += int64(size)
	s.TimingStat.Record(dur)
}

// AvgSize returns the average size.
func (s *StoreStat) AvgSize() int64 {
	if s.Count == 0 {
		return 0
	}
	return s.TotalSize / s.Count
}

// NativeStat tracks statistics for a single native operation.
type NativeStat struct {
	TimingStat
}

// Record adds a native operation sample with optional duration.
func (s *NativeStat) Record(dur time.Duration) {
	s.TimingStat.Record(dur)
}

// SubOpStat tracks statistics for a single sub-operation type.
type SubOpStat struct {
	TimingStat
}

// Record adds a sub-operation sample with optional duration.
func (s *SubOpStat) Record(dur time.Duration) {
	s.TimingStat.Record(dur)
}

// LocationStat aggregates stats by source location.
type LocationStat struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	FuncName string `json:"func,omitempty"`
	PkgPath  string `json:"pkg"`
	Count    int64  `json:"count"`
	TotalNs  int64  `json:"total_ns,omitempty"`
	Gas      int64  `json:"gas"`
}

// VarStat tracks statistics for a specific variable assignment.
type VarStat struct {
	TimingStat
	Name    string `json:"name,omitempty"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	PkgPath string `json:"pkg,omitempty"`
	Index   int    `json:"index,omitempty"`
}

// DisplayName returns a human-readable name for the variable.
func (v *VarStat) DisplayName() string {
	if v.Name != "" {
		return v.Name
	}
	if v.Index >= 0 {
		return "#" + strconv.Itoa(v.Index)
	}
	return "-"
}

// StackFrame represents a call stack frame.
type StackFrame struct {
	Func    string `json:"func"`
	PkgPath string `json:"pkg,omitempty"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
}

// StackSample is a stack sample with aggregated gas and timing.
type StackSample struct {
	Stack      []StackFrame `json:"stack"`
	Gas        int64        `json:"gas"`
	DurationNs int64        `json:"duration_ns,omitempty"`
	Count      int64        `json:"count,omitempty"`
}

// ---- Stack entries for in-progress measurements

// opStackEntry tracks an in-progress opcode measurement that was paused.
type opStackEntry struct {
	op        Op
	startTime time.Time // only used if timing enabled
	elapsed   time.Duration
	ctx       OpContext // source location context
}

// storeStackEntry tracks an in-progress store operation for nested calls.
type storeStackEntry struct {
	op        StoreOp
	startTime time.Time // only used if timing enabled
}

// nativeEntry tracks an in-progress native operation.
type nativeEntry struct {
	op        NativeOp
	startTime time.Time // only used if timing enabled
}

// subOpStackEntry tracks an in-progress sub-operation measurement.
type subOpStackEntry struct {
	op        SubOp
	startTime time.Time // only used if timing enabled
	ctx       SubOpContext
}

// ---- Call stack tracking for pprof

// callFrame represents a function on the call stack during collection.
type callFrame struct {
	funcName string
	pkgPath  string
	file     string
	line     int
}

// estimatedFrameKeySize is the expected size of a stack frame key string.
// Derived from: average func name (~20) + "@" (1) + average file path (~35) + ":" (1) + line (~4) = 61
const estimatedFrameKeySize = 64

// stackSample is the internal aggregation struct for stack samples during collection.
type stackSample struct {
	gas        int64
	durationNs int64
	count      int64
}
