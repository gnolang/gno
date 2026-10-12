package main

const X = iota

func f(x int) {}

func main() {
	f(X)
	var a = iota // ERROR "iota"
	_ = a
	f(X)
}

// Error:
// main/iota_outside_const_2.gno:9:10-14: cannot use iota outside constant declaration

// TypeCheckError:
// main/iota_outside_const_2.gno:9:10: cannot use iota outside constant declaration
