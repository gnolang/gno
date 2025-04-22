package main

func main() {
	const Huge = 1e1000
	println(Huge)
}

// Error:
// cannot convert untyped bigdec to float64: strconv.ParseFloat: parsing "1E+1000": value out of range

// TypeCheckError:
// main/files/types/overflow_a1.gno:5:10: cannot use Huge (untyped float constant 1e+1000) as float64 value in argument to built-in println (overflows)
