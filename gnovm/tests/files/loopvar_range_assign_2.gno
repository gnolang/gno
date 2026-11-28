package main

// A `for i = range` nested in a loop that declares `i` assigns the loop's
// `i`, not the outer one.

func main() {
	i := 9
	for i := 0; i < 2; i++ {
		for j := 0; j < 1; j++ {
			for i = range []int{4, 5} {
			}
		}
		println("after", i)
		if i > 0 {
			break
		}
	}
	println(i)
}

// Output:
// after 1
// 9
