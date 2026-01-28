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

// ---- Functional Options

// Option configures the profiler behavior.
type Option func(*Profiler)

// WithoutTiming disables wall-clock timing.
// Use this when you only need gas/count data without timing overhead.
func WithoutTiming() Option {
	return func(p *Profiler) { p.timingEnabled = false }
}

// WithoutStacks disables call stack tracking.
// Use this when pprof output is not needed (reduces memory overhead).
func WithoutStacks() Option {
	return func(p *Profiler) { p.stackEnabled = false }
}

// ---- Hot Path Entry Point

// R returns the global Recorder for hot-path operations.
// IMPORTANT: Must only be called within `if benchops.Enabled { ... }` guards.
// When Enabled=false (disabled builds), this returns nil and calling methods
// on nil will panic - but those code paths are eliminated at compile time.
func R() Recorder {
	return global.Load().recorder
}

// ---- Configuration Functions

// Start begins profiling with a new Profiler.
// Accepts optional functional options to configure the profiler.
// Uses CompareAndSwap to detect concurrent Start calls.
// Panics if profiling is already active.
func Start(opts ...Option) {
	p := New()

	// Apply options before starting
	for _, opt := range opts {
		opt(p)
	}

	p.Start()
	newState := &globalState{recorder: p, profiler: p}

	// CompareAndSwap ensures only one Start() succeeds. noopState is a singleton
	// that Stop() always restores, so this correctly detects concurrent access.
	if !global.CompareAndSwap(noopState, newState) {
		p.Stop() // Clean up the profiler we created
		panic("benchops: Start: profiler is already running")
	}
}

// Stop ends profiling, returns the results, and resets to noop recorder.
// Uses CompareAndSwap to detect concurrent Stop calls.
// Panics if profiling is not active.
func Stop() *Results {
	state := global.Load()
	if state.profiler == nil {
		panic("benchops: Stop: profiler is not running")
	}

	if !global.CompareAndSwap(state, noopState) {
		panic("benchops: Stop: concurrent access detected")
	}

	return state.profiler.Stop()
}

// IsRunning returns true if profiling is currently active.
func IsRunning() bool {
	return global.Load().profiler != nil
}
