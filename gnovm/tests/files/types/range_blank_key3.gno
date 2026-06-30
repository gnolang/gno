// Regression: `for _ = range slice` used to panic on nil key type in
// assertIndexTypeIsInt.
package main

func main() {
	s := []int{1, 2, 3}
	n := 0
	for _ = range s {
		n++
	}
	println(n)
}

// Output:
// 3
