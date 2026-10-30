package main

// A result beyond int32 needs a typed operand; the typed arithmetic then
// gives gc's exact value for the untyped expression.

func main() {
	var w int64 = int64('\U0010ffff') * '\U0010ffff'
	const c int64 = '\U0010ffff' * int64('\U0010ffff')
	var s int64 = int64('a') << 40
	var u uint64 = uint64('\U0010ffff') * '\U0010ffff' * '\U0010ffff'
	var f float64 = float64('\U0010ffff') * '\U0010ffff'
	println(w, c, s, u, int64(f))
}

// Output:
// 1241243320321 1241243320321 106652627894272 1382882836846149631 1241243320321
