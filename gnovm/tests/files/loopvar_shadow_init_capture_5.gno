package main

// A for-init reading a local `:=` or `var` that shadows an enclosing loop
// variable of the same name reads the local.

func main() {
	for x := 0; x < 1; x++ {
		x := 5
		for x := x + 1; x < 7; x++ {
			println("define", x)
		}
	}
	for x := 0; x < 1; x++ {
		var x = 5
		for x := x + 1; x < 7; x++ {
			println("var", x)
		}
	}
}

// Output:
// define 6
// var 6
