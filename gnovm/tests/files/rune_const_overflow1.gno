package main

// An untyped rune constant past int32 is range checked when it takes a
// type: here its default type, rune (#6389).

func main() {
	var r = '\U0010ffff' * '\U0010ffff'
	println(r)
}

// Error:
// main/rune_const_overflow1.gno:7:6-37: constant 1241243320321 overflows rune

// TypeCheckError:
// main/rune_const_overflow1.gno:7:10: cannot use '\U0010ffff' * '\U0010ffff' (untyped rune constant 1241243320321) as rune value in variable declaration (overflows)
