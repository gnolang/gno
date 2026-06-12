package main

// Adapted from Go's test/fixedbugs/bug186.go. Using iota outside a
// const declaration must be a compile-time error.

const X = iota

func f(x int) {}

func main() {
	f(X)
	var a = iota // ERROR "iota"
	_ = a
	f(X)
}

// Error:
// main/iota_outside_const_2.gno:12:10-14: cannot use iota outside constant declaration

// TypeCheckError:
// main/iota_outside_const_2.gno:12:10: cannot use iota outside constant declaration
