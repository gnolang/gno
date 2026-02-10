package main

var z = 0

var x = func() { println(y + z) }

var y = 1

func main() {
	x()
}

// Output:
// 1
