package main

// Conversions of untyped integer and rune constants are exact and range
// checked: float targets for runes, a single rounding to float32, and
// string(x) of any integer outside the int32 range is "\uFFFD".

type S string

const (
	up   = 1<<54 + 1<<30 + 1 // just above a float32 midpoint: rounds up
	tie  = 1<<54 + 1<<30     // a float32 midpoint: ties to even (down)
	tie2 = 1<<54 + 3<<30     // a float32 midpoint: ties to even (up)
	up64 = 1<<83 + 1<<30 + 1 // just above a float64 midpoint
)

func main() {
	var f float64 = 'a'
	var g float32 = -'a' * 2
	println(int64(f), int64(g))

	println(int64(float32(up)), int64(float32(-up)))
	println(int64(float32(tie)), int64(float32(-tie)))
	println(int64(float32(tie2)), int64(float32(-tie2)))
	var a float32 = up
	var b float32 = -up
	println(int64(a), int64(b))
	println(float64(up64) == float64(1<<83+1<<31), float64(-up64) == -float64(1<<83+1<<31))

	println(string(1<<64) == "\uFFFD", string(-1<<70) == "\uFFFD", S(1<<64) == "\uFFFD")
	println(string(1<<40) == "\uFFFD", string(1<<31) == "\uFFFD")
	println(string('a'+1), string(S('a'*2-'a')))
}

// Output:
// 97 -194
// 18014400656965632 -18014400656965632
// 18014398509481984 -18014398509481984
// 18014402804449280 -18014402804449280
// 18014400656965632 -18014400656965632
// true true
// true true true
// true true
// b a
