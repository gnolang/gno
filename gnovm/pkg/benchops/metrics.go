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
	Count   int64   `json:"count"`
	TotalNs int64   `json:"total_ns,omitempty"`
	MinNs   int64   `json:"min_ns,omitempty"`
	MaxNs   int64   `json:"max_ns,omitempty"`
	SumSqNs float64 `json:"sumsq_ns,omitempty"` // sum of squared durations for stddev/merge (float64 to avoid overflow)
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
	// Use float64 to avoid overflow: ns^2 for durations > 3s exceeds int64 max
	nsf := float64(ns)
	t.SumSqNs += nsf * nsf
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
	variance := t.SumSqNs/float64(t.Count) - meanNs*meanNs
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

// CSVTimingFields returns timing fields formatted for CSV export.
// Returns: [total_ns, avg_ns, stddev_ns, min_ns, max_ns]
func (t *TimingStat) CSVTimingFields() []string {
	return []string{
		strconv.FormatInt(t.TotalNs, 10),
		strconv.FormatInt(t.AvgNs(), 10),
		strconv.FormatInt(t.StdDevNs(), 10),
		strconv.FormatInt(t.MinNs, 10),
		strconv.FormatInt(t.MaxNs, 10),
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
	TotalSize    int64 `json:"total_size,omitempty"`    // Deprecated: use BytesRead/BytesWritten
	BytesRead    int64 `json:"bytes_read,omitempty"`    // Total bytes read from store
	BytesWritten int64 `json:"bytes_written,omitempty"` // Total bytes written to store
}

// Record adds a store sample with size and optional duration.
// The bytes are routed to BytesRead or BytesWritten based on the operation type.
func (s *StoreStat) Record(op StoreOp, bytes int, dur time.Duration) {
	s.TotalSize += int64(bytes) // Keep for backward compat
	if op.IsRead() {
		s.BytesRead += int64(bytes)
	} else if op.IsWrite() {
		s.BytesWritten += int64(bytes)
	}
	s.TimingStat.Record(dur)
}

// AvgSize returns the average size.
func (s *StoreStat) AvgSize() int64 {
	if s.Count == 0 {
		return 0
	}
	return s.TotalSize / s.Count
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
	Index   int    `json:"index,omitempty"` // 0-based index; -1 means "no index"
}

// DisplayName returns a human-readable name for the variable.
// Returns the variable name if set, otherwise "#N" for Index >= 0, otherwise "-".
func (v *VarStat) DisplayName() string {
	if v.Name != "" {
		return v.Name
	}
	// Index >= 0 means a valid 0-based index; -1 means "no index"
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
// Example: "handleRequest@gno.land/r/demo/boards/post.gno:42"
const estimatedFrameKeySize = 64

// stackSample is the internal aggregation struct for stack samples during collection.
type stackSample struct {
	gas        int64
	durationNs int64
	count      int64
}
