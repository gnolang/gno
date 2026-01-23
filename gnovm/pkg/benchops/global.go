//go:build gnobench

package benchops

// Enabled indicates whether benchops profiling is active.
// When built with -tags gnobench, this is true and profiling is enabled.
const Enabled = true

var global = New()

// SetGlobal sets the global profiler instance.
func SetGlobal(p *Profiler) {
	global = p
}

// Global returns the global profiler instance.
func Global() *Profiler {
	return global
}

// BeginOp starts timing for an opcode using the global profiler.
func BeginOp(op Op) {
	global.BeginOp(op)
}

// EndOp stops timing for the current opcode using the global profiler.
func EndOp() {
	global.EndOp()
}

// BeginStore starts timing for a store operation using the global profiler.
func BeginStore(op StoreOp) {
	global.BeginStore(op)
}

// EndStore stops timing for the current store operation using the global profiler.
func EndStore(size int) {
	global.EndStore(size)
}

// BeginNative starts timing for a native function using the global profiler.
func BeginNative(op NativeOp) {
	global.BeginNative(op)
}

// EndNative stops timing for the current native function using the global profiler.
func EndNative() {
	global.EndNative()
}

// Recovery resets the global profiler's internal state after a panic.
func Recovery() {
	global.Recovery()
}

// Reset clears the global profiler's collected data.
func Reset() {
	global.Reset()
}
