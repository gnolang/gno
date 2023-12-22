package main

func main() {
	x := 11
	var y float32
	y = 1 + 1<<x + 1
	println(y)
}

// Error:
// operators << and <<= not defined for float32
// *** CHECK THE ERR MESSAGES ABOVE, MAKE SURE IT'S WHAT YOU EXPECTED, DELETE THIS LINE AND RUN TEST AGAIN ***
