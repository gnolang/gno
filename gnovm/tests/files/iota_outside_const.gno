package main

const X = iota

func f(x int) {}

func main() {
	f(X)
	f(iota) // ERROR "iota"
	f(X)
}

// Error:
// main/iota_outside_const.gno:9:4-8: cannot use iota outside constant declaration

// TypeCheckError:
// main/iota_outside_const.gno:9:4: cannot use iota outside constant declaration
