// Package benchops provides profiling instrumentation for the GnoVM.
//
// # Build Tags
//
// Use -tags gnobench to enable profiling. Without this tag, the package
// compiles to zero overhead (dead code elimination via const Enabled = false).
//
// benchops profiles opcode execution, store operations (blockchain state access),
// native function calls, and sub-operations (fine-grained variable assignments).
// It supports call stack tracking for pprof flame graphs, hot spot analysis by
// source location, and multiple output formats (JSON, report, pprof, CSV).
//
// # Thread Safety
//
// Recording operations (BeginOp/EndOp, BeginStore/EndStore, etc.) are NOT
// thread-safe and must be called from a single goroutine. This matches the
// single-threaded nature of GnoVM execution.
//
// State transitions (Start/Stop) use atomics to detect accidental concurrent
// access and will panic if misused.
//
// # API
//
// Start() begins profiling, Stop() returns results and resets state.
// The profiler is single-threaded only; Start() panics if already running.
//
// Store operations can nest (e.g., GetPackage -> GetObjectSafe). The profiler
// uses a stack to correctly attribute timing to each operation.
//
// # Usage
//
//	benchops.Start()
//	defer func() {
//	    results := benchops.Stop()
//	    results.WriteJSON(os.Stdout)
//	}()
//
//	// If a VM panic may occur, recover and call Recovery() to reset
//	// internal stacks. The profiler remains running; Stop() still works.
//	defer func() {
//	    if r := recover(); r != nil {
//	        benchops.R().Recovery()
//	    }
//	}()
//	// ... VM execution ...
//
// # Options
//
// Wrap profiling calls in `if benchops.Enabled { ... }` guards. Without the
// gnobench build tag, Enabled is a const false, so the compiler eliminates
// these code paths entirely with zero runtime overhead.
//
// By default, both timing and call stack tracking are enabled.
// Use options to disable specific features:
//
//	benchops.Start()                          // Full profiling (timing + stacks)
//	benchops.Start(benchops.WithoutTiming())  // Count/gas only, no wall-clock timing
//	benchops.Start(benchops.WithoutStacks())  // No call stack tracking (no pprof output)
//
// # Output Formats
//
//	results.WriteJSON(w)          // Compact JSON
//	results.WriteReport(w)        // Human-readable summary (all entries)
//	results.WriteReportN(w, topN) // Human-readable summary (limited to topN)
//	results.WritePprof(w)         // pprof protobuf (go tool pprof compatible)
//	results.WriteCSV(w)           // CSV for spreadsheet analysis
//	results.WriteGolden(w, flags) // Deterministic output for golden tests
package benchops
