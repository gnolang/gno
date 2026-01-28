//go:build !gnobench

package benchops

// Enabled indicates whether benchops profiling is active.
// When built without -tags gnobench, this is false and all profiling is disabled.
// The compiler eliminates code paths guarded by `if benchops.Enabled { ... }`.
const Enabled = false

// ---- No-op stubs for compile-time symbol resolution
// These functions are never called at runtime (guarded by `if Enabled`)
// and will be eliminated by the compiler's dead code elimination.

// Option is a no-op type when profiling is disabled.
type Option func(*Profiler)

// WithoutTiming is a no-op when profiling is disabled.
func WithoutTiming() Option { return nil }

// WithoutStacks is a no-op when profiling is disabled.
func WithoutStacks() Option { return nil }

// R returns nil - never actually called because Enabled=false eliminates call sites.
func R() Recorder { return nil }

// Start is a no-op when profiling is disabled.
func Start(...Option) {}

// Stop is a no-op when profiling is disabled.
func Stop() *Results { return nil }

// IsRunning always returns false when profiling is disabled.
func IsRunning() bool { return false }
