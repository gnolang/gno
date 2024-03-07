package main

import "time"

func main() {
	println(1 && time.Second)
	//println(1 * time.Second)
}

// Error:
// main/debug/time_stdlibs.gno:5: operator && not defined on: Int64Kind
// *** CHECK THE ERR MESSAGES ABOVE, MAKE SURE IT'S WHAT YOU EXPECTED, DELETE THIS LINE AND RUN TEST AGAIN ***
