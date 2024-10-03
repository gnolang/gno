package main

import "strconv"

func main() {
	str := strconv.FormatFloat(64.0, byte('f'), -1, 64)
	println(str)
	str = strconv.FormatFloat(64.1, byte('f'), -1, 64)
	println(str)
	str = strconv.FormatFloat(64.0, byte('f'), 6, 64)
	println(str)
}

// Output:
// 64
// 64.1
// 64.000000
