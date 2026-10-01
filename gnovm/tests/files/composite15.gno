package main

func main() {
	a, b, c, d := 1, 2, 3, 4
	x := []int{
		a: b,
		c: d,
	}
	println(x)
}

// TypeCheckError:
// main/composite15.gno:6:3: index a must be integer constant; main/composite15.gno:7:3: index c must be integer constant

// Error:
// main/composite15.gno:6:3-4: slice/array literals may not contain non-const keys
