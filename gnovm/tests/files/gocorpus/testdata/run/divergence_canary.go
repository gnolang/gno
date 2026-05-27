// GnoOutput: divergence canary
// GoOutput:
// Divergence: stdlib-behavior: Go's builtin println writes to stderr (empty stdout); Gno's writes to stdout.

// Verifies the symmetric Gno-vs-Go run-mode triple. Without the three
// directives the test would FAIL with the auto-derived diff. With
// them: harness checks pinned outputs match actuals + the actuals
// differ. If Gno's println ever changes to match Go (→ stderr),
// gnoOutput becomes equal to goOutput and the divergence is stale —
// the harness then FAILs with "remove // GnoOutput / // GoOutput /
// // Divergence" so the blessing doesn't rot.

package main

func main() {
	println("divergence canary")
}
