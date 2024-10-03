package main

func main() {
	switch true {
	case true:
		goto here
	here:
		println("ok")
		return
	default:
		panic("should not happen")
	}
}

// Output:
// ok
