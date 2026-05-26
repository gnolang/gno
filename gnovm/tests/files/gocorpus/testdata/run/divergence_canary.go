// Divergence: stdlib-behavior: Go's builtin println writes to stderr; Gno's writes to stdout. Auto-derived Output is empty (stdout-only), but Gno emits the line — without the directive, the test would fail with an Output mismatch.

// Verifies the `// Divergence:` directive: a real Gno-vs-Go behavioral
// divergence is blessed and the test passes. B semantics check: if
// Gno's behavior ever changes to match Go (println→stderr), the
// directive becomes stale and the test FAILS — see runFiletest's
// deferred finalize for the inversion logic.

package main

func main() {
	println("divergence canary")
}
