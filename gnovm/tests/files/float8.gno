package main

import "math"

func main() {
	asVars()
	asConsts()
}

func asVars() {
	var i8 int8 = 127
	var i16 int16 = 32767
	var i32 int32 = 2147483647
	var i64 int64 = 9223372036854775807
	var i int = 9223372036854775807
	var u8 uint8 = 255
	var u16 uint16 = 65535
	var u32 uint32 = 4294967295
	var u64 uint64 = 18446744073709551615
	var f32 float32 = math.MaxFloat32
	var f64 float64 = math.MaxFloat64
	println(f32 / 2)
	println(f64 / 2)
	println((f32 - 1) + 1)
	println((f64 - 1) + 1)
	println((f32 / 2) * 2)
	println((f64 / 2) * 2)
	println(f32 - 1)
	println(f64 - 1)
	println(float32(i8))
	println(float64(i8))
	println(float32(i16))
	println(float64(i16))
	println(float32(i32))
	println(float64(i32))
	println(float32(i64))
	println(float64(i64))
	println(float32(i))
	println(float64(i))
	println(float32(u8))
	println(float64(u8))
	println(float32(u16))
	println(float64(u16))
	println(float32(u32))
	println(float64(u32))
	println(float32(u64))
	println(float64(u64))
	println(float32(f32))
	println(float64(f32))
	println(float32(f64))
	println(float64(f64))
}

func asConsts() {
	const i8 int8 = 127
	const i16 int16 = 32767
	const i32 int32 = 2147483647
	const i64 int64 = 9223372036854775807
	const i int = 9223372036854775807
	const u8 uint8 = 255
	const u16 uint16 = 65535
	const u32 uint32 = 4294967295
	const u64 uint64 = 18446744073709551615
	const f32Max float32 = math.MaxFloat32
	const f64Max float64 = math.MaxFloat64
	const f64Min = math.SmallestNonzeroFloat64
	const f32Min = math.SmallestNonzeroFloat32
	println(f32Max / 2)
	println(f64Max / 2)
	println((f32Max - 1) + 1)
	println((f64Max - 1) + 1)
	println((f32Max / 2) * 2)
	println((f64Max / 2) * 2)
	println(f32Max - 1)
	println(f64Max - 1)
	println(f64Min / 2)
	println(f32Min / 2)
	println(float32(i8))
	println(float64(i8))
	println(float32(i16))
	println(float64(i16))
	println(float32(i32))
	println(float64(i32))
	println(float32(i64))
	println(float64(i64))
	println(float32(i))
	println(float64(i))
	println(float32(u8))
	println(float64(u8))
	println(float32(u16))
	println(float64(u16))
	println(float32(u32))
	println(float64(u32))
	println(float32(u64))
	println(float64(u64))
	println(float32(f32Max))
	println(float64(f32Max))
	println(float64(f64Max))
}

// Output:
// 1.7014117e+38
// 8.988465674311579e+307
// 3.4028235e+38
// 1.7976931348623157e+308
// 3.4028235e+38
// 1.7976931348623157e+308
// 3.4028235e+38
// 1.7976931348623157e+308
// 0
// 7.006492321624085e-46
// 127
// 127
// 32767
// 32767
// 2.1474836e+09
// 2.147483647e+09
// 9.223372e+18
// 9.223372036854776e+18
// 9.223372e+18
// 9.223372036854776e+18
// 255
// 255
// 65535
// 65535
// 4.2949673e+09
// 4.294967295e+09
// 1.8446744e+19
// 1.8446744073709552e+19
// 3.4028235e+38
// 3.4028234663852886e+38
// 1.7976931348623157e+308
// 1.7014117e+38
// 8.988465674311579e+307
// 3.4028235e+38
// 1.7976931348623157e+308
// 3.4028235e+38
// 1.7976931348623157e+308
// 3.4028235e+38
// 1.7976931348623157e+308
// 127
// 127
// 32767
// 32767
// 2.1474836e+09
// 2.147483647e+09
// 9.223372e+18
// 9.223372036854776e+18
// 9.223372e+18
// 9.223372036854776e+18
// 255
// 255
// 65535
// 65535
// 4.2949673e+09
// 4.294967295e+09
// 1.8446744e+19
// 1.8446744073709552e+19
// 3.4028235e+38
// 3.4028234663852886e+38
// +Inf
// 1.7976931348623157e+308
