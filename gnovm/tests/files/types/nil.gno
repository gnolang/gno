package main

func main() {

	println(nil + nil)
}

// Error:
// main/nil.gno:5:10-19: operator + not defined on: nil

// TypeCheckError:
// main/nil.gno:5:10: invalid operation: operator + not defined on nil
