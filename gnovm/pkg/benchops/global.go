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

// WithTiming enables wall-clock timing for all operations.
// Without this option, only gas/count tracking is performed (minimal overhead).
// Use this when you need timing data for performance analysis.
func WithTiming() Option {
	return func(p *Profiler) { p.timingEnabled = true }
}

// WithStacks enables call stack tracking for pprof output.
// Stack tracking adds memory overhead for stack copies.
func WithStacks() Option {
	return func(p *Profiler) { p.stackEnabled = true }
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

// ---- Hot Path Functions (explicit Begin/End for VM main loop)

// BeginOp starts timing for an opcode using the global recorder.
func BeginOp(op Op) {
	global.Load().recorder.BeginOp(op)
}

// EndOp stops timing for the current opcode using the global recorder.
func EndOp() {
	global.Load().recorder.EndOp()
}

// SetOpContext sets the source location context for the current opcode.
func SetOpContext(ctx OpContext) {
	global.Load().recorder.SetOpContext(ctx)
}

// ---- Store Operations

// BeginStore starts tracking a store operation.
func BeginStore(op StoreOp) {
	global.Load().recorder.BeginStore(op)
}

// EndStore completes the current store operation with its size.
func EndStore(size int) {
	global.Load().recorder.EndStore(size)
}

// ---- Native Operations

// BeginNative starts tracking a native function call.
func BeginNative(op NativeOp) {
	global.Load().recorder.BeginNative(op)
}

// EndNative completes the current native function call.
func EndNative() {
	global.Load().recorder.EndNative()
}

// ---- SubOp Functions (explicit Begin/End for loops)

// BeginSubOp starts tracking a sub-operation with optional context.
// Use this with EndSubOp for explicit begin/end tracking in loops.
// Pass zero value SubOpContext{} if no context is needed.
func BeginSubOp(op SubOp, ctx SubOpContext) {
	global.Load().recorder.BeginSubOp(op, ctx)
}

// EndSubOp completes the current sub-operation.
func EndSubOp() {
	global.Load().recorder.EndSubOp()
}

// ---- Defer-Friendly Trace Functions

// TraceStore traces a store operation. Returns a closer that accepts size.
// Always records count. Only records timing if WithTiming() was used.
// Usage: defer benchops.TraceStore(benchops.StoreGetObject)(size)
func TraceStore(op StoreOp) func(size int) {
	rec := global.Load().recorder
	rec.BeginStore(op)
	return rec.EndStore
}

// TraceNative traces a native function call.
// Always records timing if WithTiming() was used.
// Usage: defer benchops.TraceNative(benchops.NativeXxx)()
func TraceNative(op NativeOp) func() {
	rec := global.Load().recorder
	rec.BeginNative(op)
	return rec.EndNative
}

// TraceSubOp traces a sub-operation with context.
// Always records count. Only records timing if WithTiming() was used.
// Usage: defer benchops.TraceSubOp(benchops.SubOpDefineVar, ctx)()
func TraceSubOp(op SubOp, ctx SubOpContext) func() {
	rec := global.Load().recorder
	rec.BeginSubOp(op, ctx)
	return rec.EndSubOp
}

// ---- Recovery

// Recovery resets the global profiler's internal state after a panic.
// No-op if profiler is not running.
func Recovery() {
	if p := global.Load().profiler; p != nil {
		p.Recovery()
	}
}

// ---- Call Stack Tracking

// PushCall pushes a function call onto the call stack.
// Called when entering a function.
func PushCall(funcName, pkgPath, file string, line int) {
	if p := global.Load().profiler; p != nil {
		p.PushCall(funcName, pkgPath, file, line)
	}
}

// PopCall pops the current function from the call stack.
// Called when returning from a function.
func PopCall() {
	if p := global.Load().profiler; p != nil {
		p.PopCall()
	}
}
