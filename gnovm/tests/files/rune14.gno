package main

// An untyped rune constant passed to println defaults to rune.

func main() {
	println('\U0010ffff' * '\U0010ffff')
}

// Error:
// main/rune14.gno:6:2-38: rune overflows target kind

// TypeCheckError:
// main/rune14.gno:6:10: cannot use '\U0010ffff' * '\U0010ffff' (untyped rune constant 1241243320321) as rune value in argument to built-in println (overflows)
