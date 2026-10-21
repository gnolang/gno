// run

// Transcribed from Go's go/constant tests and issues #6389 #6392 #6393 #6394; expected output from go1.25.9.
// Arithmetic on two untyped rune constants is exact: the product of two
// max runes is 1241243320321, range checked only once it takes a type.

package main

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	var w int64 = '\U0010ffff' * '\U0010ffff'
	var x int64 = c
	var y int64 = 'a' * 1000000000 // untyped rune * untyped int
	println(w, x, y)

	var a, b, d int64 = c + '\U0010ffff', 'a' - c, c * c / '\U0010ffff'
	println(a, b, d)

	var e, f, g, h int64 = c & (c - 1), c | '\U0010ffff', c ^ 'a', c &^ '\U0010ffff'
	println(e, f, g, h)

	println(c > 0, c < c*2, c == -2228223)

	// Cancellation back into rune range: gc accepts the rune 1114111.
	var r rune = c / '\U0010ffff'
	var m int64 = c % 1000
	println(r, m)

	// In-range operations on untyped runes.
	println('a'+'b', 'z'-'a', 'a'*'b', 'z'/'a', 'z'%'a', 'a'<<1, 'z'>>1)
	println('a' < 'b', 'a' == 97, 'a'+1 == 'b')
}

// GnoOutput:
// -2228223 -2228223 97000000000
// -1114112 2228320 -4
// -2228224 -2162689 -2228128 -3276800
// false false true
// -2 -223
// 195 25 9506 1 25 194 61
// true true true

// GoOutput:
// 1241243320321 1241243320321 97000000000
// 1241244434432 -1241243320224 1382882836846149631
// 1241243320320 1241243385855 1241243320416 1241242271744
// true true false
// 1114111 321
// 195 25 9506 1 25 194 61
// true true true

// Verdict: KnownIssue run: #6389 untyped rune constant arithmetic wraps at int32
