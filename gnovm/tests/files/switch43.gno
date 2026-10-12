package main

func main() {
	x := 1
	switch x {
	case 1:
		println("one")
		fallthrough
	default:
		println("default")
	case 2:
		println("two")
	}
}

// Output:
// one
// default
