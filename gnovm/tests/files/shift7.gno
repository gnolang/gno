package main

// A compound shift keeps its count's type. Counts below come from runtime
// narrowing conversions, which must shift by the converted value only.

type Bits uint8

func main() {
	var wide int64 = 264 // uint8(wide) == 8
	n := uint8(wide)

	a := int64(7)
	a <<= n
	println(a, int64(7)<<n)

	b := int64(1 << 20)
	b >>= n
	println(b)

	var w8 int64 = 261 // int8(w8) == 5
	c := uint32(1)
	c <<= int8(w8)
	println(c)

	var neg int64 = -251 // int8(neg) == 5
	d := int64(-64)
	d >>= int8(neg)
	println(d)

	var w16 int64 = 1<<16 + 2 // uint16(w16) == 2
	e := uint8(3)
	e <<= uint16(w16)
	println(e)

	var w32 uint64 = 1<<32 + 3 // uint32(w32) == 3
	f := uint64(5)
	f <<= uint32(w32)
	println(f)

	// Arithmetic on the narrowed value keeps its type.
	g := int64(3)
	g <<= n - 4
	println(g)

	// So does a declared integer type.
	h := 1
	h <<= Bits(wide)
	println(h)

	// A bound checked on the narrowed value holds for the shift too.
	var dec int64 = 260 // uint8(dec) == 4
	if d := uint8(dec); d <= 4 {
		price := int64(1_000_000)
		price >>= d
		println(price)
	}

	var w16b int64 = 1<<16 + 5 // int16(w16b) == 5
	i := int64(1)
	i <<= int16(w16b)
	println(i)

	var w32b int64 = 1<<32 + 5 // int32(w32b) == 5
	j := int64(1)
	j <<= int32(w32b)
	println(j)

	k := int64(1)
	k <<= rune(w32b)
	println(k)
}

// Output:
// 1792 1792
// 4096
// 32
// -2
// 12
// 40
// 48
// 256
// 62500
// 32
// 32
// 32
