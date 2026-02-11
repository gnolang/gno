//go:build !gnobench

package benchops

// Enabled is false when built without -tags gnobench.
// All profiling calls are eliminated by the compiler's dead code elimination.
const Enabled = false

type Option func(*Profiler)

func WithoutTiming() Option    { return nil }
func WithoutStacks() Option    { return nil }
func R() Recorder              { return noopRecorder{} }
func Start(...Option)          {}
func Stop() *Results           { return nil }
func IsRunning() bool          { return false }
