package main

// A `for _, v = range` nested in a loop that declares `v` assigns the loop's
// `v`, not the outer one.

func main() {
	v := 99
	for k, v := 0, 0; k < 2; k++ {
		for _, v = range []int{k + 10} {
		}
		println(v)
	}
	println(v)
}

// Output:
// 10
// 11
// 99
