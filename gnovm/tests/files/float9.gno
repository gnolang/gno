// Issue 7419: odd behavior for float constants underflowing to 0.
// Go silently rounds such constants to 0 when typed as float64.

package main

func main() {
	var x = 1e-779137
	var y = 1e-779138
	if x != 0 {
		panic("x != 0")
	}
	if y != 0 {
		panic("y != 0")
	}
	println("ok")
}

// Output:
// ok
