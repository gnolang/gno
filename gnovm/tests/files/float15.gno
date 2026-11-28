// Untyped integer constants beyond int64 range convert to float targets
// (constant conversion rounds; it must not take the default int path).
package main

func main() {
	a := float64(1 << 100)
	b := float32(1 << 100)
	c := float64(-(1 << 100))
	d := float64(1<<100 + 1) // rounds to nearest
	println(a)
	println(b)
	println(c)
	println(d)
}

// Output:
// 1.2676506002282294e+30
// 1.2676506e+30
// -1.2676506002282294e+30
// 1.2676506002282294e+30
