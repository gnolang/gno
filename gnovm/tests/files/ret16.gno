package main

func assertWithClosure(x any) func() {
	return func() {
		if _, ok := x.(string); ok {
			println("is captured:", ok, x)
		}
	}
}

func main() {
	assertWithClosure("a")()
}

// Output:
// is captured: true a
