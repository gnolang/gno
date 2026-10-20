package main

// gc accepts this and prints 1241243320321; Gno rejects it (see rune32).

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	var y int64 = big << s
	println(y)
}

// Error:
// main/rune33.gno:10:16-24: rune overflows target kind
