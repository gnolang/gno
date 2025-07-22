package main

func init() {
	println("here")
}

func main() {
	init()
}

// Error:
// main/init1.gno:8:2-6: name init not declared

// TypeCheckError:
// main/init1.gno:8:2: undefined: init
