package main

func main() {
	var a, b uint256 = 1, 0
	var c, d int256 = -1, 0
	try(func() { println(a / b) })
	try(func() { println(a % b) })
	try(func() { println(c / d) })
	try(func() { println(c % d) })
	try(func() { a /= b })
	try(func() { c %= d })
}

func try(f func()) {
	defer func() {
		println("recovered:", recover())
	}()
	f()
}

// Output:
// recovered: runtime error: division by zero
// recovered: runtime error: division by zero
// recovered: runtime error: division by zero
// recovered: runtime error: division by zero
// recovered: runtime error: division by zero
// recovered: runtime error: division by zero
