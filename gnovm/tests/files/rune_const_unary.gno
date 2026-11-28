package main

// Unary - + ^ on an untyped rune constant, a rune shifted past int32 and a
// rune converted to a float type: exact, kind kept (#6389).

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	var x int64 = -'a'
	var y int64 = ^'a'
	var z int64 = -c
	var w rune = -'a'
	var v rune = ^'a'
	println(x, y, z, w, v)
	println(-'a', ^'a', +'a', -(c / '\U0010ffff'))
	var s int64 = 'a' << 40
	var f float64 = 'a'
	var g float32 = -'a'
	println(s, int64(f), int64(g), float64('a') == 97, float32(c) == 1241243320321)
	println(-1, ^1, ^0, +1, ^uint8(0), ^uint32(0), ^int8(0), -int8(5))
}

// Output:
// -97 -98 -1241243320321 -97 -98
// -97 -98 97 -1114111
// 106652627894272 97 -97 true true
// -1 -2 -1 1 255 4294967295 -1 -5
