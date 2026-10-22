// run

// Transcribed from Go's go/constant tests and issue #6389; expected output from go1.25.9.
// A constant shift of an untyped rune is exact in Go: the result is range
// checked only once it takes a type.

package main

func main() {
	var x int64 = 'a' << 40
	println(x)
}

// GnoOutput:

// GnoError:
// main/rune_shift_const.go:10:16-25: constant overflows

// GoOutput:
// 106652627894272

// Verdict: KnownIssue run: #6389 untyped rune constant shift overflows int32
