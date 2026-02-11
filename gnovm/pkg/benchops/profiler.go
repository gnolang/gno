package benchops

import (
	"sync/atomic"
	"time"
)

const (
	// maxOpCodes is the maximum number of opcodes (byte range 0-255).
	maxOpCodes = 256
	// defaultStackCapacity is the initial capacity for op/store stacks.
	defaultStackCapacity = 16
	// defaultStackSampleCapacity is the initial capacity for stack samples.
	defaultStackSampleCapacity = 256
)

// Recorder is the interface for recording profiling measurements.
// The global recorder is either a noopRecorder (when not profiling) or
// a *Profiler (when profiling is active).
//
// Error handling semantics:
//   - EndOp, EndStore: panic if called without matching Begin (programming error)
//   - EndSubOp: tolerant of missing BeginSubOp (returns silently) to simplify conditional instrumentation
type Recorder interface {
	// BeginOp starts timing for an opcode with its source location context.
	BeginOp(op Op, ctx OpContext)
	// EndOp stops timing for the current opcode and records the measurement.
	EndOp()
	// BeginStore starts timing for a store operation.
	BeginStore(op StoreOp)
	// EndStore stops timing for the current store operation with the given size.
	EndStore(size int)
	// TraceNative returns a function to be deferred for tracing native calls.
	TraceNative(op NativeOp) func()
	// BeginSubOp starts timing for a sub-operation within an opcode.
	BeginSubOp(op SubOp, ctx SubOpContext)
	// EndSubOp stops timing for the current sub-operation (tolerant of missing Begin).
	EndSubOp()
	// PushCall pushes a function call onto the call stack for pprof tracking.
	PushCall(funcName, pkgPath, file string, line int)
	// PopCall pops the current function from the call stack.
	PopCall()
	// Recovery resets internal state after a panic without changing profiler state.
	Recovery()
}

// noopRecorder is the default recorder that does nothing.
// Used when profiling is not active (before Start() in enabled builds).
type noopRecorder struct{}

// noopFunc is a package-level no-op function to avoid allocation in TraceNative.
var noopFunc = func() {}

func (noopRecorder) BeginOp(Op, OpContext)                {}
func (noopRecorder) EndOp()                               {}
func (noopRecorder) BeginStore(StoreOp)                   {}
func (noopRecorder) EndStore(int)                         {}
func (noopRecorder) TraceNative(NativeOp) func()          { return noopFunc }
func (noopRecorder) BeginSubOp(SubOp, SubOpContext)       {}
func (noopRecorder) EndSubOp()                            {}
func (noopRecorder) PushCall(string, string, string, int) {}
func (noopRecorder) PopCall()                             {}
func (noopRecorder) Recovery()                            {}

// Ensure noopRecorder implements Recorder.
var _ Recorder = noopRecorder{}

// Ensure Profiler implements Recorder
var _ Recorder = (*Profiler)(nil)

// State represents the profiler's current state.
type State int32

const (
	StateIdle    State = iota // Not started, ready for Start()
	StateRunning              // Actively profiling
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	default:
		return "unknown"
	}
}

func (p *Profiler) loadState() State                    { return State(p.state.Load()) }
func (p *Profiler) storeState(s State)                   { p.state.Store(int32(s)) }
func (p *Profiler) casState(old, new State) bool {
	return p.state.CompareAndSwap(int32(old), int32(new))
}

// Profiler collects timing statistics for GnoVM operations.
// Recording (BeginOp/EndOp, etc.) is NOT thread-safe and must be done
// from a single goroutine. Atomics are only used for state transitions
// to detect accidental concurrent Start/Stop calls.
type Profiler struct {
	startTime time.Time
	stopTime  time.Time

	opStack        []opStackEntry
	currentOp      *opStackEntry
	storeStack     []storeStackEntry
	currentNative  *nativeEntry
	locationStats  map[string]*LocationStat
	currentSubOp   *subOpStackEntry        // sub-ops don't nest, single pointer suffices
	varStats       map[string]*VarStat     // key: "file:line:varname" or "file:line:idx"
	callStack      []callFrame             // current call stack (root-to-leaf)
	stackSampleAgg map[string]*stackSample // aggregated samples by stack signature (avoids memory growth)

	state atomic.Int32

	timingEnabled bool
	stackEnabled  bool

	opStats     [maxOpCodes]OpStat
	storeStats  [maxOpCodes]StoreStat
	nativeStats [maxOpCodes]TimingStat
	subOpStats  [maxSubOps]TimingStat
}

// New creates a new Profiler in idle state.
// By default, timing and stack tracking are enabled for full profiling.
// Use WithoutTiming() or WithoutStacks() to disable specific features.
func New() *Profiler {
	p := &Profiler{
		timingEnabled:  true, // Enabled by default for full profiling
		stackEnabled:   true, // Enabled by default for pprof output
		opStack:        make([]opStackEntry, 0, defaultStackCapacity),
		storeStack:     make([]storeStackEntry, 0, defaultStackCapacity),
		locationStats:  make(map[string]*LocationStat),
		varStats:       make(map[string]*VarStat),
		callStack:      make([]callFrame, 0, defaultStackCapacity),
		stackSampleAgg: make(map[string]*stackSample, defaultStackSampleCapacity),
	}
	p.storeState(StateIdle)
	return p
}

// Start begins profiling. Panics if not in StateIdle.
func (p *Profiler) Start() {
	if !p.casState(StateIdle, StateRunning) {
		if p.loadState() == StateRunning {
			panic("benchops: Start: profiler is already running (concurrent access or missing Stop)")
		}
		panic("benchops: Start: profiler is not idle (invalid state)")
	}

	// Clear any stale data from instrumentation that ran without Start/Stop
	p.clearData()

	p.startTime = time.Now()
}

// Stop ends profiling, returns the results, and resets to idle state.
// Panics if not in StateRunning.
func (p *Profiler) Stop() *Results {
	if !p.casState(StateRunning, StateIdle) {
		panic("benchops: Stop: profiler is not running (missing Start)")
	}

	p.stopTime = time.Now()
	results := p.buildResults()

	// Clear data for reuse (state already transitioned to idle)
	p.clearData()

	return results
}

// clearData clears all collected measurement data without changing state.
func (p *Profiler) clearData() {
	p.opStats = [maxOpCodes]OpStat{}
	p.opStack = p.opStack[:0]
	p.currentOp = nil
	p.storeStats = [maxOpCodes]StoreStat{}
	p.storeStack = p.storeStack[:0]
	p.nativeStats = [maxOpCodes]TimingStat{}
	p.currentNative = nil
	p.locationStats = make(map[string]*LocationStat)
	p.subOpStats = [maxSubOps]TimingStat{}
	p.currentSubOp = nil
	p.varStats = make(map[string]*VarStat)
	p.callStack = p.callStack[:0]
	p.stackSampleAgg = make(map[string]*stackSample)
}

// Reset clears all collected data and returns to idle state.
// This is a no-op if the profiler is already idle.
// Panics if called while the profiler is running (use Stop() instead).
func (p *Profiler) Reset() {
	if p.loadState() == StateRunning {
		panic("benchops: Reset: profiler is running (use Stop instead)")
	}
	p.clearData()
	p.storeState(StateIdle)
}

// Recovery resets internal state after a panic without changing profiler state.
// Call this from a recover block to ensure the profiler can continue.
func (p *Profiler) Recovery() {
	p.opStack = p.opStack[:0]
	p.currentOp = nil
	p.storeStack = p.storeStack[:0]
	p.currentNative = nil
	p.currentSubOp = nil
	p.callStack = p.callStack[:0]
}

// State returns the current profiler state.
// This is safe to call at any time (lock-free atomic read).
func (p *Profiler) State() State {
	return p.loadState()
}
