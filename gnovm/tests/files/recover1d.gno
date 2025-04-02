package main

func main() {
	defer func() {
		recover()
		println("all recovered")
	}()

	defer func() {
		recover()
		panic("other panic")
	}()

	panic("test panic")
}

// Output:
// all recovered
