package main

// An exact untyped rune constant beyond int32 next to a non-constant
// shift, in an int64 context.

var s uint

const big = '\U0010ffff' * '\U0010ffff'

func main() {
	var y int64 = ('a' << s) + big
	println(y)
}

// Output:
// 1241243320418
