package main

// Outer var with same name as loop var should not be affected by the rename.
// Inside the loop, i refers to i.loopvar (the loop var).
// After the loop, i refers to the outer var (99).
func main() {
	i := 99
	for i := 0; i < 3; i++ {
		println("loop:", i)
	}
	println("outer:", i)
}

// Output:
// loop: 0
// loop: 1
// loop: 2
// outer: 99
