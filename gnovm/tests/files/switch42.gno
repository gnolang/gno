package main

func main() {
	x := 99
	switch x {
	default:
		println("default")
		fallthrough
	case 1:
		println("one")
	case 2:
		println("two")
	}
}

// Output:
// default
// one
