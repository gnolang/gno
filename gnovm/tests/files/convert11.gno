package main

// Per the Go spec, converting an integer to string yields the UTF-8 of the
// code point; values outside the valid Unicode code point range yield "�".
// Values that do not fit in an int32 must NOT be truncated first.
func main() {
	// 64-bit values that would wrap to valid code points if truncated to int32.
	println(string(0x100000041)) // NOT "A"
	println(string(-4294967231)) // NOT "A" (== -(1<<32) + 65)
	println(string(0x10001F600)) // NOT an emoji
	var i64 int64 = 0x100000041
	println(string(i64))
	var i int = 0x100000041
	println(string(i))
	var u64 uint64 = 0x100000041
	println(string(u64))
	var u uint = 18446744073709551615
	println(string(u))
	// in-range values still convert normally.
	println(string(rune(65)), string(int64(0x1F600)), string(uint64(65)))
	// invalid in-range values yield "�".
	println(string(-1), string(0xD800), string(0x110000))
	// uint32 uses the unchanged conversion path: values > MaxInt32
	// reinterpret as negative runes and (0x10FFFF, MaxInt32] is out of
	// range, so every invalid uint32 already yields "�".
	println(string(uint32(0x80000041)), string(uint32(0x7FFFFFFF)))
	// untyped rune constants fold through the same conversion.
	println(string('A' + 0x100000000))
}

// Output:
// �
// �
// �
// �
// �
// �
// �
// A 😀 A
// � � �
// � �
// �
