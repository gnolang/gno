// MAXALLOC: 500000000
package main

// Assigns a [1000]int to the blank identifier ten times. Nothing reads "_",
// so the array is not copied; the right-hand side is still evaluated.
func get(a *[1000]int) [1000]int {
	println("get")
	return *a
}

func main() {
	var a [1000]int
	for i := 0; i < 10; i++ {
		_ = a
	}
	_ = get(&a)
}

// Output:
// get

// Gas:
// 160201
