package main

func main() {

	println(nil + nil)
}

// Error:
// main/files/types/nil.gno:5:10: operator + not defined on: nil

// TypeCheckError:
// main/files/types/nil.gno:5:10: invalid operation: operator + not defined on nil
