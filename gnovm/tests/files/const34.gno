package main

const a = func() { println("hey") }

func main() {
	println("ok")
}

// Error:
// main/const34.gno:3:7-36: func func(){ println<VPInvalid(0)>("hey") } (variable of type func()) is not constant

// TypeCheckError:
// main/const34.gno:3:11: (func() literal) (value of type func()) is not constant
