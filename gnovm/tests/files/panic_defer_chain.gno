// Minimal defer-repanic chain.
// Asserts the per-panic stacktrace + joined Error survive markAbort/Run.
// Pre-PR (UnhandledPanicError) used to drop the per-Exception Stacktrace
// on the floor; this test pins that both panics' frames are rendered.

package main

func main() {
	defer func() { panic("B") }()
	panic("A")
}

// Stacktrace:
// panic: A
// main<VPBlock(1,0)>()
//     main/panic_defer_chain.gno:10
// panic: B
// defer func(){ ... }()
//     main/panic_defer_chain.gno:9
// main<VPBlock(1,0)>()
//     main/panic_defer_chain.gno:9

// Error:
// A
// 	B
