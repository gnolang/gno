//go:build gnobench

package benchops

import "sync/atomic"

// Enabled is true when built with -tags gnobench.
const Enabled = true

type globalState struct {
	recorder Recorder
	profiler *Profiler // non-nil when profiling is active
}

var global atomic.Pointer[globalState]

var noopState = &globalState{recorder: noopRecorder{}}

func init() {
	global.Store(noopState)
}

// ---- Options

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

// R returns the global Recorder. When profiling is not active, returns a
// noop recorder. Call within `if benchops.Enabled { ... }` guards for
// compile-time dead code elimination in non-profiling builds.
func R() Recorder {
	return global.Load().recorder
}

// Start begins profiling with a new Profiler.
// Panics if profiling is already active.
func Start(opts ...Option) {
	p := New()

	// Apply options before starting
	for _, opt := range opts {
		opt(p)
	}

	p.Start()
	newState := &globalState{recorder: p, profiler: p}

	if !global.CompareAndSwap(noopState, newState) {
		p.Stop() // Clean up the profiler we created
		panic("benchops: Start: profiler is already running")
	}
}

// Stop ends profiling, returns the results, and resets to noop recorder.
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
