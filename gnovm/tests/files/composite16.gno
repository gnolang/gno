package main

func x(n int) int {
	println(n)
	return n
}

func main() {
	x := []int{
		x(1): x(2),
		x(3): x(4),
	}
	println(x)
}

// Output:
// 1
// 2
// 3
// 4
// slice[(0 int),(2 int),(0 int),(4 int)]

// TypeCheckError:
// main/files/composite16.gno:10:3: index x(1) must be integer constant; main/files/composite16.gno:11:3: index x(3) must be integer constant
