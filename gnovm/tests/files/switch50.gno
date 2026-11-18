package main

// The switch block is re-created per iteration, but its clause slots are also
// reused within one execution by each fallthrough.
func main() {
	for i := 0; i < 3; i++ {
		switch i {
		case 0:
			s := "zero"
			println(s)
			fallthrough
		case 1:
			n := i * 10
			println(n)
			fallthrough
		case 2:
			println("tail", i)
		}
	}
}

// Output:
// zero
// 0
// tail 0
// 10
// tail 1
// tail 2
