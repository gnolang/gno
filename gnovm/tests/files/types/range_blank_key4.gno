// Blank key and value over array, in the assign (not define) form.
package main

func main() {
	a := [3]int{1, 2, 3}
	n := 0
	for _, _ = range a {
		n++
	}
	println(n)
}

// Output:
// 3
