package main

// As rune20, next to a typed variable.

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	var x int64
	println(x + big)
}

// Output:
// 1241243320321
