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
)

// Recorder is the interface for recording profiling measurements.
// The global recorder is either a noopRecorder (when not profiling) or
// a *Profiler (when profiling is active).
type Recorder interface {
	BeginOp(op Op)
	SetOpContext(ctx OpContext) // Set source location context for current op
	EndOp()
	BeginStore(op StoreOp)
	EndStore(size int)
	BeginNative(op NativeOp)
	EndNative()
}

// noopRecorder is the default recorder that does nothing.
// Used when profiling is not active.
type noopRecorder struct{}

func (noopRecorder) BeginOp(Op)             {}
func (noopRecorder) SetOpContext(OpContext) {}
func (noopRecorder) EndOp()                 {}
func (noopRecorder) BeginStore(StoreOp)     {}
func (noopRecorder) EndStore(int)           {}
func (noopRecorder) BeginNative(NativeOp)   {}
func (noopRecorder) EndNative()             {}

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

	// Op statistics (gas-only by default)
	opStats      [maxOpCodes]opStat
	opStatsTimed [maxOpCodes]opStatTimed // only populated if timingEnabled
	opStack      []opStackEntry
	currentOp    *opStackEntry

	// Store statistics
	storeStats      [maxOpCodes]storeStat
	storeStatsTimed [maxOpCodes]storeStatTimed // only populated if timingEnabled
	storeStack      []storeStackEntry

	// Native statistics
	nativeStats      [maxOpCodes]nativeStat
	nativeStatsTimed [maxOpCodes]nativeStatTimed // only populated if timingEnabled
	currentNative    *nativeEntry

	// Location statistics (key: "file:line")
	locationStats map[string]*locationStat
}

// New creates a new Profiler in idle state.
func New() *Profiler {
	p := &Profiler{
		opStack:       make([]opStackEntry, 0, defaultStackCapacity),
		storeStack:    make([]storeStackEntry, 0, defaultStackCapacity),
		locationStats: make(map[string]*locationStat),
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
	p.opStats = [maxOpCodes]opStat{}
	p.opStatsTimed = [maxOpCodes]opStatTimed{}
	p.opStack = p.opStack[:0]
	p.currentOp = nil
	p.storeStats = [maxOpCodes]storeStat{}
	p.storeStatsTimed = [maxOpCodes]storeStatTimed{}
	p.storeStack = p.storeStack[:0]
	p.nativeStats = [maxOpCodes]nativeStat{}
	p.nativeStatsTimed = [maxOpCodes]nativeStatTimed{}
	p.currentNative = nil
	p.locationStats = make(map[string]*locationStat)
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
}

// State returns the current profiler state.
// This is safe to call at any time (lock-free atomic read).
func (p *Profiler) State() State {
	return State(p.state.Load())
}
