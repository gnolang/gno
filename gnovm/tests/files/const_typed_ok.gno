package main

// Typed integer constants that fit keep their exact value. Non-constant
// arithmetic still wraps, which is what the last three lines show.
func main() {
	const c int8 = 100
	println(c + 27)

	const n8 int8 = -64
	println(n8 << 1)

	const s8 int8 = -1
	println(s8 >> 8)
	println(s8 >> 100)

	const m8 int8 = -128
	println(m8 >> 10)
	println(m8 % -1)
	println(m8 / 2)

	const q int8 = -5
	println(q / 2)

	const c16 int16 = 32767
	println(c16 - 1)

	const c32 int32 = 1
	println(c32 << 30)

	const c64 int64 = 1
	println(c64 << 62)

	const neg64 int64 = -1
	println(neg64 >> 64)
	println(neg64 >> 65)

	const u8 uint8 = 200
	println(u8 + 55)

	const u16 uint16 = 1
	println(u16 << 15)

	const u32 uint32 = 1
	println(u32 << 31)

	const u64 uint64 = 1
	println(u64 << 63)
	println(u64 >> 100)

	const ui uint = 3
	println(ui << 4)

	const i int = 1
	println(i << 4)

	println(int8(127))
	println(uint8(255))
	println(int8(int16(100)))

	const neg int8 = -127
	println(-neg)

	const z uint8 = 0
	println(-z)

	const by byte = 200
	println(by + 50)

	const rn rune = 1000
	println(rn * 2)

	const maxInt64 int64 = (1 << 63) - 1
	println(maxInt64)

	const maxUint64 uint64 = 1<<64 - 1
	println(maxUint64)

	var a int8 = 100
	println(a * 2)

	var x, y uint64 = 8, 40
	println(x - y)

	var g int32 = -2147483648
	println(-g)
}

// Output:
// 127
// -128
// -1
// -1
// -1
// 0
// -64
// -2
// 32766
// 1073741824
// 4611686018427387904
// -1
// -1
// 255
// 32768
// 2147483648
// 9223372036854775808
// 0
// 48
// 16
// 127
// 255
// 100
// 127
// 0
// 250
// 2000
// 9223372036854775807
// 18446744073709551615
// -56
// 18446744073709551584
// -2147483648
