package main

func main() {
	println(-3.0 * 1.0 / 2.0) // -1.5
	println(-3.0 * 1.0 / 2)   // -1.5
	println(-3.0 * 1 / 2.0)   // -1.5

	println(-3 * 1.0 / 2.0) // 1.5
	println(-3 * 1.0 / 2)   // 1.5
	println(-3 * 1 / 2.0)   // 1.5
}

// Output:
// -1.5
// -1.5
// -1.5
// -1.5
// -1.5
// -1.5
