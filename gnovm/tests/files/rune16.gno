package main

// An untyped rune constant may exceed int64 but not when converted to it.

func main() {
	var x int64 = '\U0010ffff' * '\U0010ffff' * '\U0010ffff' * '\U0010ffff'
	println(x)
}

// Error:
// main/rune16.gno:6:6-73: rune overflows target kind

// TypeCheckError:
// main/rune16.gno:6:16: cannot use '\U0010ffff' * '\U0010ffff' * '\U0010ffff' * '\U0010ffff' (untyped rune constant 1540684980241500611543041) as int64 value in variable declaration (overflows)
