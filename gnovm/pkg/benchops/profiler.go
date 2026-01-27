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
// - EndOp, EndStore, EndNative: panic if called without matching Begin (programming error)
// - EndSubOp: tolerant of missing BeginSubOp (returns silently) to simplify conditional instrumentation
type Recorder interface {
	BeginOp(op Op)
	SetOpContext(ctx OpContext)
	EndOp()
	BeginStore(op StoreOp)
	EndStore(size int)
	BeginNative(op NativeOp)
	EndNative()
	BeginSubOp(op SubOp, ctx SubOpContext)
	EndSubOp()
}

// noopRecorder is the default recorder that does nothing.
// Used when profiling is not active.
type noopRecorder struct{}

func (noopRecorder) BeginOp(Op)                     {}
func (noopRecorder) SetOpContext(OpContext)         {}
func (noopRecorder) EndOp()                         {}
func (noopRecorder) BeginStore(StoreOp)             {}
func (noopRecorder) EndStore(int)                   {}
func (noopRecorder) BeginNative(NativeOp)           {}
func (noopRecorder) EndNative()                     {}
func (noopRecorder) BeginSubOp(SubOp, SubOpContext) {}
func (noopRecorder) EndSubOp()                      {}

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

// Profiler collects timing statistics for GnoVM operations.
// Recording (BeginOp/EndOp, etc.) is NOT thread-safe and must be done
// from a single goroutine. Atomics are only used for state transitions
// to detect accidental concurrent Start/Stop calls.
type Profiler struct {
	state     atomic.Int32
	startTime time.Time
	stopTime  time.Time

	// Configuration (set via options before Start)
	timingEnabled bool // Track wall-clock time (expensive, opt-in)
	stackEnabled  bool // Track call stacks for pprof (opt-in)

	// Op statistics (unified type handles both timed and non-timed)
	opStats   [maxOpCodes]OpStat
	opStack   []opStackEntry
	currentOp *opStackEntry

	// Store statistics
	storeStats [maxOpCodes]StoreStat
	storeStack []storeStackEntry

	// Native statistics
	nativeStats   [maxOpCodes]NativeStat
	currentNative *nativeEntry

	// Location statistics (key: "file:line")
	locationStats map[string]*LocationStat

	// Sub-operation statistics
	subOpStats   [maxSubOps]SubOpStat
	currentSubOp *subOpStackEntry    // sub-ops don't nest, single pointer suffices
	varStats     map[string]*VarStat // key: "file:line:varname" or "file:line:idx"

	// Call stack tracking for pprof (opt-in)
	callStack      []callFrame             // current call stack (root-to-leaf)
	stackSampleAgg map[string]*stackSample // aggregated samples by stack signature (avoids memory growth)
}

// New creates a new Profiler in idle state.
func New() *Profiler {
	p := &Profiler{
		opStack:        make([]opStackEntry, 0, defaultStackCapacity),
		storeStack:     make([]storeStackEntry, 0, defaultStackCapacity),
		locationStats:  make(map[string]*LocationStat),
		varStats:       make(map[string]*VarStat),
		callStack:      make([]callFrame, 0, defaultStackCapacity),
		stackSampleAgg: make(map[string]*stackSample, defaultStackSampleCapacity),
	}
	p.state.Store(int32(StateIdle))
	return p
}

// Start begins profiling.
// Uses atomic CompareAndSwap to detect concurrent access - panics if profiler
// is already running.
// Clears any stale data from previous instrumentation before starting.
// Panics if not in StateIdle.
func (p *Profiler) Start() {
	if !p.state.CompareAndSwap(int32(StateIdle), int32(StateRunning)) {
		current := State(p.state.Load())
		if current == StateRunning {
			panic("benchops: profiler is already running (concurrent access or missing Stop)")
		}
		panic("benchops: Start called on non-idle profiler (invalid state)")
	}

	// Clear any stale data from instrumentation that ran without Start/Stop
	p.clearData()

	p.startTime = time.Now()
}

// Stop ends profiling, returns the results, and resets to idle state.
// The profiler can be immediately reused with Start() after Stop().
// Uses atomic CompareAndSwap to ensure thread-safe state transition.
// Panics if not in StateRunning.
func (p *Profiler) Stop() *Results {
	if !p.state.CompareAndSwap(int32(StateRunning), int32(StateIdle)) {
		panic("benchops: Stop called on non-running profiler (missing Start)")
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
	p.nativeStats = [maxOpCodes]NativeStat{}
	p.currentNative = nil
	p.locationStats = make(map[string]*LocationStat)
	p.subOpStats = [maxSubOps]SubOpStat{}
	p.currentSubOp = nil
	p.varStats = make(map[string]*VarStat)
	p.callStack = p.callStack[:0]
	p.stackSampleAgg = make(map[string]*stackSample)
}

// Reset clears all collected data and returns to idle state.
// This is a no-op if the profiler is already idle.
// Panics if called while the profiler is running (use Stop() instead).
func (p *Profiler) Reset() {
	current := State(p.state.Load())
	if current == StateRunning {
		panic("benchops: Reset called on running profiler (use Stop() instead)")
	}
	p.clearData()
	p.state.Store(int32(StateIdle))
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
	return State(p.state.Load())
}
