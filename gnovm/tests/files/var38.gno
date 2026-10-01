// Two chained multi-value decls — both must be split correctly.
// var A, B = C, D and var C, D = E, F are both one-to-one multi-value decls.
package main

var (
	A, B = C, D
	C, D = E, F
	E    = 7
	F    = 8
)

func main() {
	println(A, B, C, D) // 7 8 7 8
}

// Output:
// 7 8 7 8
