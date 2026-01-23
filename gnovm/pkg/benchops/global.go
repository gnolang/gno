//go:build gnobench

package benchops

import "sync/atomic"

// Enabled indicates whether benchops profiling is active.
// When built with -tags gnobench, this is true and profiling is enabled.
const Enabled = true

// globalState holds the current recorder and optional profiler reference.
type globalState struct {
	recorder Recorder
	profiler *Profiler // non-nil only when profiling is active
}

var global atomic.Pointer[globalState]

// noopState is the singleton idle state used when not profiling.
var noopState = &globalState{recorder: noopRecorder{}}

func init() {
	global.Store(noopState)
}

// Start begins profiling with a new Profiler.
// Uses CompareAndSwap to detect concurrent Start calls.
// Panics if profiling is already active.
func Start() {
	p := New()
	p.Start()
	newState := &globalState{recorder: p, profiler: p}

	if !global.CompareAndSwap(noopState, newState) {
		p.Stop() // Clean up the profiler we created
		panic("benchops: Start called while profiler is already running")
	}
}

// Stop ends profiling, returns the results, and resets to noop recorder.
// Uses CompareAndSwap to detect concurrent Stop calls.
// Panics if profiling is not active.
func Stop() *Results {
	state := global.Load()
	if state.profiler == nil {
		panic("benchops: Stop called while profiler is not running")
	}

	if !global.CompareAndSwap(state, noopState) {
		panic("benchops: concurrent Stop detected")
	}

	return state.profiler.Stop()
}

// IsRunning returns true if profiling is currently active.
func IsRunning() bool {
	return global.Load().profiler != nil
}

// BeginOp starts timing for an opcode using the global recorder.
func BeginOp(op Op) {
	global.Load().recorder.BeginOp(op)
}

// EndOp stops timing for the current opcode using the global recorder.
func EndOp() {
	global.Load().recorder.EndOp()
}

// BeginStore starts timing for a store operation using the global recorder.
func BeginStore(op StoreOp) {
	global.Load().recorder.BeginStore(op)
}

// EndStore stops timing for the current store operation using the global recorder.
func EndStore(size int) {
	global.Load().recorder.EndStore(size)
}

// BeginNative starts timing for a native function using the global recorder.
func BeginNative(op NativeOp) {
	global.Load().recorder.BeginNative(op)
}

// EndNative stops timing for the current native function using the global recorder.
func EndNative() {
	global.Load().recorder.EndNative()
}

// Recovery resets the global profiler's internal state after a panic.
// No-op if profiler is not running.
func Recovery() {
	if p := global.Load().profiler; p != nil {
		p.Recovery()
	}
}
