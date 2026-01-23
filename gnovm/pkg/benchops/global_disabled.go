//go:build !gnobench

package benchops

// Enabled indicates whether benchops profiling is active.
// When built without -tags gnobench, this is false and all profiling is disabled.
// The compiler eliminates code paths guarded by `if benchops.Enabled { ... }`.
const Enabled = false
