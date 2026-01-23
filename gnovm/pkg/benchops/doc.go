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
// # Store Call Nesting
//
// Store operations can nest (e.g., GetPackage -> GetObjectSafe -> GetPackageRealm).
// The profiler uses a stack (not depth counter) to correctly attribute timing
// to each operation. BeginStore automatically pauses opcode timing on the first
// call; EndStore resumes it when the stack empties.
//
// # Panic Recovery
//
// If the VM panics during execution, call Profiler.Recovery() to reset internal
// state. This allows profiling to continue after the panic is handled. Note that
// Recovery() does not release the mutex - the profiler remains in running state.
// You must still call Stop() to release the lock and get results.
//
// # Thread Safety and Atomic State
//
// The profiler is designed for SINGLE-THREADED use only. Machines should NOT
// run in parallel when profiling. The profiler enforces this with fail-fast
// behavior:
//
//   - Start() uses atomic CompareAndSwap and panics if profiler is already running
//   - Measurement methods (BeginOp, EndOp, etc.) have minimal checks for clear error messages
//   - Misuse (e.g., EndOp without BeginOp) causes an immediate panic with clear message
//   - State is stored atomically for lock-free reads via State()
//
// This design means:
//
//   - Low overhead in the measurement hot path
//   - Concurrent access is detected immediately via panic
//   - Programming errors (misuse) fail fast with clear messages
//
// # Simple State Machine
//
// The profiler has only two states: Idle and Running.
//
//	Idle  --Start()--> Running --Stop()--> Idle
//
// Stop() automatically resets the profiler to Idle state, so it can be
// immediately reused without calling Reset(). The Reset() method exists
// for explicit clearing if needed, but is typically unnecessary.
//
// # Test Usage Pattern
//
// Example test with profiling:
//
//	func TestWithProfiling(t *testing.T) {
//	    // Do NOT call t.Parallel() - profiler is single-threaded
//
//	    p := benchops.Global()
//	    p.Start()  // Panics if another test is profiling
//	    defer func() {
//	        results := p.Stop()  // Auto-resets to Idle
//	        results.WriteJSON(os.Stdout)
//	    }()
//
//	    // ... test code with Machine ...
//	}
//
// # Usage
//
//	runtime.GOMAXPROCS(1) // For accurate measurements
//	p := benchops.Global()
//	p.Start()
//	defer func() {
//	    results := p.Stop()
//	    results.WriteJSON(os.Stdout)
//	}()
//	// ... VM execution ...
package benchops
