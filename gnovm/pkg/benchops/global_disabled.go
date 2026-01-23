//go:build !gnobench

package benchops

// Enabled indicates whether benchops profiling is active.
// When built without -tags gnobench, this is false and all profiling is disabled.
// The compiler eliminates code paths guarded by `if benchops.Enabled { ... }`.
const Enabled = false

// ---- No-op stubs for compile-time symbol resolution
// These functions are never called at runtime (guarded by `if Enabled`)
// and will be eliminated by the compiler's dead code elimination.

func Start()          {}
func Stop() *Results  { return nil }
func IsRunning() bool { return false }
func BeginOp(op Op)   {}
func EndOp()                  {}
func BeginStore(op StoreOp)   {}
func EndStore(size int)       {}
func BeginNative(op NativeOp) {}
func EndNative()              {}
func Recovery()               {}
