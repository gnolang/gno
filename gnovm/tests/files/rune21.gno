package main

// As rune20, shifted by a variable.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	var y int64 = big << s
	println(y)
}

// Output:
// 1241243320321
