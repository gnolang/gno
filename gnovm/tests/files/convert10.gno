// Test case where a slice of a user-defined byte type (not uint8 or byte) is
// converted to a string.  Same for slice of runes.
// Regression test for issue #5776.

package main

type MyByte byte

type MyRune rune

func main() {
	var y []MyByte
	_ = string(y)

	var z []MyRune
	_ = string(z)
}
