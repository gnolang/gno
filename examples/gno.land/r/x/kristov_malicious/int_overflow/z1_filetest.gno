package main

import "gno.land/r/x/kristov_malicious/int_overflow"

func main() {
	println(overflow.Render(""))
}

// Output:
// -9223372036854775808 vs 9223372036854775808
