// Package benchops provides profiling instrumentation for the GnoVM.
//
// # Build Tags
//
// Use -tags gnobench to enable profiling. Without this tag, the package
// compiles to zero overhead (dead code elimination via const Enabled = false).
//
// # API
//
// Start() begins profiling, Stop() returns results and resets state.
// The profiler is single-threaded only; Start() panics if already running.
//
// Store operations can nest (e.g., GetPackage -> GetObjectSafe). The profiler
// uses a stack to correctly attribute timing to each operation.
//
// Call Recovery() after a VM panic to reset internal state before continuing.
// The profiler remains running; call Stop() to get results.
//
// # Usage
//
//	benchops.Start()
//	defer func() {
//	    results := benchops.Stop()
//	    results.WriteJSON(os.Stdout)
//	}()
//	// ... VM execution ...
package benchops
