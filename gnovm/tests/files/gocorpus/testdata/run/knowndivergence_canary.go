// Verifies the `// Verdict: KnownDivergence run: <reason>` flow: Gno
// and Go disagree, the file says so and why, and verify passes as
// long as both pinned outputs still hold. The divergence used is
// permanent by design: the Go spec leaves the output format of the
// builtin print/println implementation-defined, and Gno formats
// floats the way fmt's %v does where gc prints +1.500000e+000.

package main

func main() {
	println(1.5)
	println(float32(2))
	println(uint8(200), int64(-3))
}

// GnoOutput:
// 1.5
// 2
// 200 -3

// GoOutput:
// +1.500000e+000
// +2.000000e+000
// 200 -3

// Verdict: KnownDivergence run: builtin println float format is implementation-defined (Go spec); Gno uses fmt %v formatting
