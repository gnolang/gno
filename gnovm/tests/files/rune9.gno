package main

// Unary operators on untyped rune constants are exact.

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	var a int64 = -c
	var b int64 = -'\U0010ffff' * '\U0010ffff'
	var x int64 = ^c
	var p int64 = +c
	var m int64 = -(-2147483647 - 'a') // -(untyped rune -2147483744)
	println(a, b, x, p, m)
	r := -'a' // in range, defaults to rune
	println([]rune{r}[0], ^'a')
}

// Output:
// -1241243320321 -1241243320321 -1241243320322 1241243320321 2147483744
// -97 -98
