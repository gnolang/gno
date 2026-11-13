package main

// A compound untyped switch tag with a non-constant shift keeps master's
// typing: the cases match for every arithmetic and bitwise operator, in
// both operand orders.

var s uint

func main() {
	switch 1 + (1 << s) {
	case 2:
		println("1 + (1 << s)", "match")
	default:
		println("1 + (1 << s)", "miss")
	}
	switch (1 << s) + 1 {
	case 2:
		println("(1 << s) + 1", "match")
	default:
		println("(1 << s) + 1", "miss")
	}
	switch 1 - (1 << s) {
	case 0:
		println("1 - (1 << s)", "match")
	default:
		println("1 - (1 << s)", "miss")
	}
	switch (1 << s) - 1 {
	case 0:
		println("(1 << s) - 1", "match")
	default:
		println("(1 << s) - 1", "miss")
	}
	switch 1 * (1 << s) {
	case 1:
		println("1 * (1 << s)", "match")
	default:
		println("1 * (1 << s)", "miss")
	}
	switch (1 << s) * 1 {
	case 1:
		println("(1 << s) * 1", "match")
	default:
		println("(1 << s) * 1", "miss")
	}
	switch 1 / (1 << s) {
	case 1:
		println("1 / (1 << s)", "match")
	default:
		println("1 / (1 << s)", "miss")
	}
	switch (1 << s) / 1 {
	case 1:
		println("(1 << s) / 1", "match")
	default:
		println("(1 << s) / 1", "miss")
	}
	switch 1 % (1 << s) {
	case 0:
		println("1 % (1 << s)", "match")
	default:
		println("1 % (1 << s)", "miss")
	}
	switch (1 << s) % 1 {
	case 0:
		println("(1 << s) % 1", "match")
	default:
		println("(1 << s) % 1", "miss")
	}
	switch 1 & (1 << s) {
	case 1:
		println("1 & (1 << s)", "match")
	default:
		println("1 & (1 << s)", "miss")
	}
	switch (1 << s) & 1 {
	case 1:
		println("(1 << s) & 1", "match")
	default:
		println("(1 << s) & 1", "miss")
	}
	switch 1 | (1 << s) {
	case 1:
		println("1 | (1 << s)", "match")
	default:
		println("1 | (1 << s)", "miss")
	}
	switch (1 << s) | 1 {
	case 1:
		println("(1 << s) | 1", "match")
	default:
		println("(1 << s) | 1", "miss")
	}
	switch 1 ^ (1 << s) {
	case 0:
		println("1 ^ (1 << s)", "match")
	default:
		println("1 ^ (1 << s)", "miss")
	}
	switch (1 << s) ^ 1 {
	case 0:
		println("(1 << s) ^ 1", "match")
	default:
		println("(1 << s) ^ 1", "miss")
	}
	switch 1 &^ (1 << s) {
	case 0:
		println("1 &^ (1 << s)", "match")
	default:
		println("1 &^ (1 << s)", "miss")
	}
	switch (1 << s) &^ 1 {
	case 0:
		println("(1 << s) &^ 1", "match")
	default:
		println("(1 << s) &^ 1", "miss")
	}
}

// Output:
// 1 + (1 << s) match
// (1 << s) + 1 match
// 1 - (1 << s) match
// (1 << s) - 1 match
// 1 * (1 << s) match
// (1 << s) * 1 match
// 1 / (1 << s) match
// (1 << s) / 1 match
// 1 % (1 << s) match
// (1 << s) % 1 match
// 1 & (1 << s) match
// (1 << s) & 1 match
// 1 | (1 << s) match
// (1 << s) | 1 match
// 1 ^ (1 << s) match
// (1 << s) ^ 1 match
// 1 &^ (1 << s) match
// (1 << s) &^ 1 match
