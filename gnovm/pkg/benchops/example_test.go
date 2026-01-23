//go:build gnobench

package benchops_test

import (
	"bytes"
	"fmt"
	"os"

	"github.com/gnolang/gno/gnovm/pkg/benchops"
)

// Example_globalAPI demonstrates the typical usage pattern with the global API.
// Start() begins profiling, Stop() returns results and resets to noop state.
func Example_globalAPI() {
	// Start profiling (panics if already running)
	benchops.Start()

	// Simulate some operations
	benchops.BeginOp(benchops.OpAdd)
	benchops.EndOp()

	benchops.BeginOp(benchops.OpCall)
	// Store operations are nested and auto-pause opcode timing
	benchops.BeginStore(benchops.StoreGetObject)
	benchops.EndStore(100) // size in bytes
	benchops.EndOp()

	// Stop profiling and get results
	results := benchops.Stop()

	// Print human-readable report to stdout
	results.WriteReport(os.Stdout, 10)
}

// Example_profilerDirect demonstrates direct Profiler usage without globals.
// This pattern is useful for embedded profiling or custom scenarios.
func Example_profilerDirect() {
	// Create a new profiler instance
	p := benchops.New()
	p.Start()

	// Record operations
	p.BeginOp(benchops.OpAdd)
	p.EndOp()

	p.BeginOp(benchops.OpMul)
	p.EndOp()

	// Get results
	results := p.Stop()

	// Verify we captured the operations
	if results.OpStats["OpAdd"] != nil {
		fmt.Println("Captured OpAdd")
	}
	if results.OpStats["OpMul"] != nil {
		fmt.Println("Captured OpMul")
	}
	// Output:
	// Captured OpAdd
	// Captured OpMul
}

// Example_storeNesting demonstrates how store operations nest correctly.
// Each BeginStore/EndStore pair tracks time independently.
func Example_storeNesting() {
	p := benchops.New()
	p.Start()

	// Start an opcode
	p.BeginOp(benchops.OpCall)

	// Nested store calls (like GetPackage -> GetObject -> GetPackageRealm)
	p.BeginStore(benchops.StoreGetPackage)
	p.BeginStore(benchops.StoreGetObject) // Second level
	p.EndStore(100)                       // End GetObject
	p.EndStore(200)                       // End GetPackage

	p.EndOp()

	results := p.Stop()

	// Both store operations were captured
	if results.StoreStats["StoreGetPackage"] != nil {
		fmt.Println("Captured StoreGetPackage")
	}
	if results.StoreStats["StoreGetObject"] != nil {
		fmt.Println("Captured StoreGetObject")
	}
	// Output:
	// Captured StoreGetPackage
	// Captured StoreGetObject
}

// Example_resultsJSON demonstrates JSON output for machine processing.
func Example_resultsJSON() {
	p := benchops.New()
	p.Start()

	p.BeginOp(benchops.OpAdd)
	p.EndOp()

	results := p.Stop()

	// Write JSON to buffer
	var buf bytes.Buffer
	results.WriteJSON(&buf)

	// JSON contains the OpStats
	output := buf.String()
	if len(output) > 0 {
		fmt.Println("JSON output generated")
	}
	// Output:
	// JSON output generated
}

// Example_panicRecovery demonstrates how to recover from panics during profiling.
// After a panic, call Recovery() to reset internal stacks, then continue or stop.
func Example_panicRecovery() {
	p := benchops.New()
	p.Start()

	// Simulate partial operations (like a panic mid-execution)
	p.BeginOp(benchops.OpCall)
	p.BeginStore(benchops.StoreGetPackage)
	// ... panic would occur here ...

	// Recovery resets the stacks but keeps profiler running
	p.Recovery()

	// Can continue profiling after recovery
	p.BeginOp(benchops.OpAdd)
	p.EndOp()

	results := p.Stop()

	// Only completed operations are in results
	if results.OpStats["OpAdd"] != nil {
		fmt.Println("Captured post-recovery OpAdd")
	}
	// Output:
	// Captured post-recovery OpAdd
}
