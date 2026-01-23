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
// state. This prevents corrupted timing data on subsequent runs.
//
// # Usage
//
//	p := benchops.New(benchops.DefaultConfig())
//	benchops.SetGlobal(p)
//	p.Start()
//	defer func() {
//	    results := p.Stop()
//	    results.WriteJSON(os.Stdout)
//	}()
//	// ... VM execution ...
package benchops
