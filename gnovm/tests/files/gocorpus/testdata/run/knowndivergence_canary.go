// Verifies the `// Verdict: KnownDivergence run: <reason>` flow: Gno
// and Go disagree on purpose, the file pins both sides and cites the
// Gno decision it rests on. Verify passes while both outputs hold and
// fails if Gno ever starts matching Go ("stale verdict").
//
// The decision: MaxTypeDepth caps composite type nesting at 8 so a
// type expression cannot be used to exhaust the preprocessor; Go has
// no such cap and runs this program.

package main

func main() {
	var x [][][][][][][][][]int
	println(len(x))
}

// GnoOutput:

// GnoError:
// main/knowndivergence_canary.go:13:8-29: type [][][][][][][][][](const-type int) nesting depth 9 exceeds max 8

// GoOutput:
// 0

// Verdict: KnownDivergence run: MaxTypeDepth caps composite type nesting at 8, a preprocess hardening cap (#5642)
