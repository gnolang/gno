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

func WithTiming() Option                                { return nil }
func WithStacks() Option                                { return nil }
func Start(opts ...Option)                              {}
func Stop() *Results                                    { return nil }
func IsRunning() bool                                   { return false }
func BeginOp(op Op)                                     {}
func SetOpContext(ctx OpContext)                        {}
func EndOp()                                            {}
func BeginStore(op StoreOp)                             {}
func EndStore(size int)                                 {}
func BeginNative(op NativeOp)                           {}
func EndNative()                                        {}
func BeginSubOp(op SubOp, ctx SubOpContext)             {}
func EndSubOp()                                         {}
func TraceStore(op StoreOp) func(int)                   { return func(int) {} }
func TraceNative(op NativeOp) func()                    { return func() {} }
func TraceSubOp(op SubOp, ctx SubOpContext) func()      { return func() {} }
func Recovery()                                         {}
func PushCall(funcName, pkgPath, file string, line int) {}
func PopCall()                                          {}
