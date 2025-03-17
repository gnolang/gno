package main

type zilch any

func main() {
	_ = zilch(nil)

	println("ok")
}

// Output:
// ok
