// run

// Transcribed from Go's go/constant tests and issues #6389 #6392 #6393 #6394; expected output from go1.25.9.
// Comparisons and arithmetic on big untyped constants are exact; the
// operand tables of go/constant's value_test.go, in filetest form.

package main

const big = 1 << 100

func main() {
	println(big > 1<<99, big == 1<<100, big < 1<<101, big-1 < big)
	println(big/(1<<98), big>>90, big%(1<<99) == 0, (big+1)%2)
	println(1<<63 > 1<<62, -(1<<63) < 0, 1<<64-1 > 1<<63, 1<<64-1 == 18446744073709551615)
	println(uint64(1<<64-1) > 1<<63, int64(1<<63-1) == 9223372036854775807)

	println(0 == 0, 0 != 0, 0 < 10, 10 <= 10, 0 > 10, 10 >= 10)
	println(1.0/123456789 == 1.0/123456789, 1.0/123456789 < 1.0/123456788, 1.0/123456788 <= 1.0/123456789)
	println(0.11 > 0.11, 0.11 >= 0.11, 0.1+0.9 == 1, 1e100+1e100 == 2e100, 1e100-1e100 == 0)
	println("foo" == "bar", "foo" != "bar", "foo" < "bar", "foo" <= "bar", "foo" > "bar", "foo" >= "bar")
	println(false == false, false == true, true != false)

	println(10/2, 5/3, 10%3, 0xff&0xf, 0xb|0xa0, 1^-1, 0xf&^1, 1&^0xf)
	println(1<<10, 1024>>10, 0<<0, 0>>0, 5.0/3 > 1.6, 5.0/3 < 1.7)
	println(1<<100 > 3.5, 1e30 == 1<<99+1, 2e2 == 200, 'a' < 1000)
	println(123456789012345678901234567890/1000000000000000000000000000, 123456789012345678901234567890%1000000000000000000000000000 == 456789012345678901234567890)
}

// GnoOutput:
// true true true true
// 4 1024 true 1
// true true true true
// true true
// true false true true false true
// true true false
// false true true true true
// false true false false true true
// true false true
// 5 1 1 15 171 -2 14 0
// 1024 1 0 0 true true
// true false true true
// 123 true

// GoOutput:
// true true true true
// 4 1024 true 1
// true true true true
// true true
// true false true true false true
// true true false
// false true true true true
// false true false false true true
// true false true
// 5 1 1 15 171 -2 14 0
// 1024 1 0 0 true true
// true false true true
// 123 true

// Verdict: Clean run
