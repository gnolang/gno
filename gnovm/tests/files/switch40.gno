package main

func main() {
	a := 1
	switch c := int16(a); c {
	case 0:
		panic("should not happen")
	case 1:
		println("done")
	default:
		panic("should not happen")
	}
}

// Output:
// done
