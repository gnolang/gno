// Regression: `for _, v = range slice` used to panic on nil key type in
// assertIndexTypeIsInt.
package main

func main() {
	a := []int{1, 2, 3}
	var v int
	for _, v = range a {
	}
	println(v)
}

// Output:
// 3
