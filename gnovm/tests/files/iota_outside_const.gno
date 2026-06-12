package main

// Adapted from Go's test/fixedbugs/bug186.go. Using iota outside a
// const declaration must be a compile-time error.

const X = iota

func f(x int) {}

func main() {
	f(X)
	f(iota) // ERROR "iota"
	f(X)
}

// Error:
// main/iota_outside_const.gno:12:4-8: cannot use iota outside constant declaration

// TypeCheckError:
// main/iota_outside_const.gno:12:4: cannot use iota outside constant declaration
