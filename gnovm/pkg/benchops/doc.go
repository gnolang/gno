// Package benchops provides profiling instrumentation for the GnoVM.
//
// # Build Tags
//
// The package uses a single build tag for conditional compilation:
//
//   - gnobench: Enable profiling (all measurements active)
//
// Without the gnobench tag, benchops compiles to zero overhead. This is
// achieved through a const Enabled = false that the compiler uses to
// eliminate dead code. No function calls, no nil checks, no interface
// dispatch - the instrumentation code is completely removed from the binary.
//
// # Zero Overhead Guarantee
//
// Production builds (without -tags gnobench) have absolutely no performance
// impact. The compiler eliminates all benchops code paths when:
//
//	if benchops.Enabled {
//	    benchops.BeginOp(op)  // This entire block is removed
//	}
//
// Verify with: go tool nm <binary> | grep benchops
// Should show only type/const symbols, no function symbols.
//
// # Noop Recorder Pattern
//
// When profiling is not active, BeginOp/EndOp/etc. delegate to a noop
// recorder that does nothing. When Start() is called, a real profiler is
// created and set as the global recorder. Stop() returns results and
// resets back to the noop recorder. This avoids any "stale data" issues.
//
// # Store Call Nesting
//
// Store operations can nest (e.g., GetPackage -> GetObjectSafe -> GetPackageRealm).
// The profiler uses a stack (not depth counter) to correctly attribute timing
// to each operation. BeginStore automatically pauses opcode timing on the first
// call; EndStore resumes it when the stack empties.
//
// # Panic Recovery
//
// If the VM panics during execution, call Recovery() to reset internal
// state. This allows profiling to continue after the panic is handled.
// The profiler remains in running state - you must still call Stop() to
// get results.
//
// # Thread Safety
//
// The profiler is designed for SINGLE-THREADED use only. Machines should NOT
// run in parallel when profiling. The profiler enforces this with fail-fast
// behavior:
//
//   - Start() panics if profiler is already running
//   - Misuse (e.g., EndOp without BeginOp) causes an immediate panic with clear message
//
// This design means:
//
//   - Low overhead in the measurement hot path (interface dispatch only)
//   - Concurrent access is detected immediately via panic
//   - Programming errors (misuse) fail fast with clear messages
//
// # Simple State Machine
//
// The global recorder is either a noop (not profiling) or a real profiler (profiling).
//
//	noop  --Start()--> profiler --Stop()--> noop
//
// Stop() returns results and automatically resets to noop state.
//
// # Test Usage Pattern
//
// Example test with profiling:
//
//	func TestWithProfiling(t *testing.T) {
//	    // Do NOT call t.Parallel() - profiler is single-threaded
//
//	    benchops.Start()  // Panics if another test is profiling
//	    defer func() {
//	        results := benchops.Stop()
//	        results.WriteJSON(os.Stdout)
//	    }()
//
//	    // ... test code with Machine ...
//	}
//
// # Usage
//
//	runtime.GOMAXPROCS(1) // For accurate measurements
//	benchops.Start()
//	defer func() {
//	    results := benchops.Stop()
//	    results.WriteJSON(os.Stdout)
//	}()
//	// ... VM execution ...
package benchops
