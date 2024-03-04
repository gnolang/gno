package main

func main() {
	a := "hello"
	a++
	println(a)
}

// Error:
// main/debug/inc.gno:5: operator ++ not defined on: StringKind
