// Package instrumentation defines the low-level observability hooks exposed by
// the Gno VM. The goal is to mirror the flexibility of eBPF perf events and
// Go's runtime/pprof sample API without forcing higher-level profilers to link
// against VM internals.
package instrumentation

// FrameSnapshot describes a single frame in a call stack at the moment a
// sample is captured. This structure intentionally mirrors runtime/pprof's
// location metadata so that profilers can export compatible data formats.
type FrameSnapshot struct {
	FuncName string
	File     string
	PkgPath  string
	Line     int
	Column   int
	Inline   bool
	IsCall   bool
}

// SampleContext encapsulates VM execution state required to build CPU/gas
// samples. It is inspired by the perf sample records used by eBPF programs and
// the go tool pprof profile format.
type SampleContext struct {
	Frames  []FrameSnapshot
	Cycles  int64
	GasUsed int64
	// MachineID is a best-effort identifier for the originating machine so
	// profilers can separate baselines when multiple machines share a sink.
	MachineID uintptr
}

// AllocationEvent represents a heap allocation attributed to a specific stack.
type AllocationEvent struct {
	Bytes   int64
	Objects int64
	Kind    string
	Stack   []FrameSnapshot
	// MachineID carries the originating machine for attribution, when available.
	MachineID uintptr
}

// LineSample captures per-line execution metrics similar to pprof's line-level
// reports (e.g. "go tool pprof -lines").
type LineSample struct {
	Func   string
	File   string
	Line   int
	Cycles int64
	// MachineID carries the originating machine for attribution, when available.
	MachineID uintptr
}

// Sink receives instrumentation events emitted by the VM. Implementations may
// choose to ignore callbacks they do not care about; for production profiling
// code this interface is typically satisfied by a struct inside
// gnovm/pkg/profiler.
type Sink interface {
	OnSample(*SampleContext)
	OnAllocation(*AllocationEvent)
	OnLineSample(*LineSample)
}

// Capabilities allows sinks to declare which event types they need so the VM
// can avoid unnecessary work when instrumentation is disabled.
type Capabilities interface {
	WantsSamples() bool
	WantsAllocations() bool
	WantsLineSamples() bool
}
