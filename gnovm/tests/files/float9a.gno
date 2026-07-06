// A negative exponent too large for int32 is a malformed constant in Go,
// not an underflow-to-0: it must stay an error.

package main

func main() {
	var d = 1e-9999999999999999999
	println(d == 0)
}

// Error:
// main/float9a.gno:7:10-32: invalid decimal constant: 1e-9999999999999999999

// TypeCheckError:
// main/float9a.gno:7:10: malformed constant: 1e-9999999999999999999
