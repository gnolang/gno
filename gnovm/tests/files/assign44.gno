package main

func f() {}

func main() {
	f = nil
}

// Error:
// main/assign44.gno:6:2-9: cannot assign to func f

// TypeCheckError:
// main/assign44.gno:6:2: cannot assign to f (neither addressable nor a map index expression)
