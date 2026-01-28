// Package benchops provides profiling instrumentation for the GnoVM.
//
// # Build Tags
//
// Use -tags gnobench to enable profiling. Without this tag, the package
// compiles to zero overhead (dead code elimination via const Enabled = false).
//
// # Features
//
//   - Opcode execution tracking with optional timing
//   - Store operation profiling (blockchain state access)
//   - Native function call tracking
//   - Sub-operation profiling (fine-grained variable assignments)
//   - Call stack tracking for pprof flame graphs
//   - Hot spot analysis by source location
//   - Multiple output formats: JSON, human-readable report, pprof, CSV
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
// # Performance Characteristics
//
// When built without -tags gnobench:
//   - Zero overhead (all calls compile to no-ops via const Enabled = false)
//
// When built with -tags gnobench:
//   - Default (full profiling): ~50ns per BeginOp/EndOp pair (timing + stack tracking)
//   - WithoutTiming: ~10ns per pair (count/gas only)
//   - WithoutStacks: Reduced memory overhead (no stack aggregation)
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
//
// # Options
//
// By default, both timing and call stack tracking are enabled for full profiling.
// Use options to disable specific features:
//
//	benchops.Start()                          // Full profiling (timing + stacks)
//	benchops.Start(benchops.WithoutTiming())  // Disable wall-clock timing (gas/count only)
//	benchops.Start(benchops.WithoutStacks())  // Disable call stack tracking (no pprof)
//
// # Output Formats
//
//	results.WriteJSON(w)          // Compact JSON
//	results.WriteReport(w, topN)  // Human-readable summary
//	results.WritePprof(w)         // pprof protobuf (go tool pprof compatible)
//	results.WriteCSV(w)           // CSV for spreadsheet analysis
//	results.WriteGolden(w, flags) // Deterministic output for golden tests
package benchops
