// run

// Transcribed from Go's go/constant tests and issues #6389 #6392 #6393 #6394; expected output from go1.25.9.
// Unary - and ^ on untyped rune and int constants are exact and keep the
// operand's kind: -'a' is the untyped rune constant -97. An untyped rune
// constant also converts to a float type like any integer constant.

package main

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	var x int64 = -'a'
	var y int64 = ^'a'
	var z int64 = -c
	var w rune = -'a'
	var v rune = ^'a'
	println(x, y, z, w, v)
	println(-'a', ^'a', +'a', -'a'*2, ^'a'+1)

	println(-1, ^1, ^0, +1, -(1 << 62), ^(1 << 62))
	println(-0.5 < 0, -1e100 < -1e99, -(1<<100) < -(1<<99))
	println(^uint8(0), ^uint32(0), ^int8(0), -int8(5), ^int64(0))
	println(!true, !(1 < 2), !false)

	var f float64 = 'a'
	var g float32 = -'a' * 2
	println(int64(f), int64(g))

	const k = -'a' * 2
	var kk int64 = k
	var r rune = -(c / '\U0010ffff')
	println(kk, r)
}

// GnoOutput:

// GnoError:
// main/unary_const.go:13:16-20: unexpected type <untyped> int32 in operation

// GoOutput:
// -97 -98 -1241243320321 -97 -98
// -97 -98 97 -194 -97
// -1 -2 -1 1 -4611686018427387904 -4611686018427387905
// true true true
// 255 4294967295 -1 -5 -1
// false false true
// 97 -194
// -194 -1114111

// Verdict: KnownIssue run: #6389 unary - and ^ on, and float conversion of, an untyped rune constant are rejected
