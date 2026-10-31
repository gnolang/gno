package main

// Arithmetic on untyped rune constants is exact (Go spec, Constant
// expressions): the product of two max runes is 1241243320321, range
// checked only once it takes a type (#6389).

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	var w int64 = '\U0010ffff' * '\U0010ffff'
	var x int64 = c
	var y int64 = 'a' * 1000000000
	println(w, x, y)
	var a, b, d int64 = c + '\U0010ffff', 'a' - c, c * c / '\U0010ffff'
	println(a, b, d)
	var e, f, g, h int64 = c & (c - 1), c | '\U0010ffff', c ^ 'a', c &^ '\U0010ffff'
	println(e, f, g, h)
	println(c > 0, c < c*2, c == -2228223)
	var r rune = c / '\U0010ffff'
	var m int64 = c % 1000
	var u uint64 = c
	var fl float64 = c
	println(r, m, u, int64(fl))
	println('a'+'b', 'z'-'a', 'a'*'b', 'z'/'a', 'z'%'a', 'a'<<1, 'z'>>1)
}

// Output:
// 1241243320321 1241243320321 97000000000
// 1241244434432 -1241243320224 1382882836846149631
// 1241243320320 1241243385855 1241243320416 1241242271744
// true true false
// 1114111 321 1241243320321 1241243320321
// 195 25 9506 1 25 194 61
