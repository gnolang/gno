package main

func main() {
	const Huge = 1e1000
	println(Huge)
}

// Error:
// cannot convert untyped bigdec to float64: strconv.ParseFloat: parsing "1E+1000": value out of range
