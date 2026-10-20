package main

// Untyped constant arithmetic is exact: an operation on two untyped runes
// is not computed in int32, even though its operands are runes.

const big = '\U0010ffff' * '\U0010ffff' // untyped rune 1241243320321

func main() {
	var mul int64 = '\U0010ffff' * '\U0010ffff'
	var add int64 = big + '\U0010ffff'
	var sub int64 = 'a' - big
	var quo int64 = big * big / '\U0010ffff'
	var rem int64 = big * 'a' % 2147483647
	var and int64 = big & (big - 1)
	var or int64 = big | '\U0010ffff'
	var xor int64 = big ^ 'a'
	var andnot int64 = big &^ '\U0010ffff'
	println(mul, add, sub, quo, rem, and, or, xor, andnot)

	var u uint64 = '\U0010ffff' * '\U0010ffff' * '\U0010ffff'
	var f float64 = '\U0010ffff' * '\U0010ffff'
	var f32 float32 = '\U0010ffff' * '\U0010ffff'
	// Floats are printed as integers so the output is the same under gc.
	println(u, int64(f), int64(f32))

	// Explicit conversions take the same exact value.
	println(int64(big), uint64(big), int64(float64(big)))
}

// Output:
// 1241243320321 1241244434432 -1241243320224 1382882836846149631 1931402082 1241243320320 1241243385855 1241243320416 1241242271744
// 1382882836846149631 1241243320321 1241243320320
// 1241243320321 1241243320321 1241243320321
