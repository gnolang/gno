package main

func main() {
	a, b, c, d := 1, 2, 3, 4
	x := []int{
		a: b,
		c: d,
	}
	println(x)
}

// Output:
// slice[(0 int),(2 int),(0 int),(4 int)]

// TypeCheckError:
// main/composite15.gno:6:3: index a must be integer constant; main/composite15.gno:7:3: index c must be integer constant
