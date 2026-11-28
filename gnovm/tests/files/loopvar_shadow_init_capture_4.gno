package main

// A for-init reading a parameter that shadows an enclosing loop variable of
// the same name reads the parameter.

func main() {
	for x := 0; x < 2; x++ {
		func(x int) {
			for x := x + 1; x < 12; x++ {
				println("p", x)
			}
		}(10)
	}
}

// Output:
// p 11
// p 11
