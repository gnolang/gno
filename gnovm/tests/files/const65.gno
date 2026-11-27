package main

const k = 1

// A composite-literal key before a shadowing declaration in the same block
// refers to the outer const.
func main() {
	{
		a := []int{k: 7}
		println(len(a), a[1])
		b := [3]int{k: 7}
		println(len(b), b[1])
		k := 5
		_ = k
	}
}

// Output:
// 2 7
// 3 7
