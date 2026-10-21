// run

// Transcribed from Go's go/constant tests and issues #6389 #6392 #6393 #6394; expected output from go1.25.9.
// A non-constant shift of an untyped constant takes the type the constant
// would have if the shift were replaced by its left operand alone, so the
// type comes from the context of the whole expression (#6393), and an
// untyped rune op an untyped int yields an untyped rune (#6339).

package main

func show(v interface{}) {
	_, ok := v.(rune)
	println(ok)
}

func main() {
	var s uint = 32
	var s0 uint = 0

	println(int64(('a' << s) + ('b' << s)))
	var x1 int64 = ('a' << s) + ('b' << s)
	println(x1)
	println(int64(('a' << s) + (1 << s)))
	var x2 int64 = ('a' << s) + (1 << s)
	println(x2)
	println(('a' << s) < (1 << s))

	x3 := 1 + ('a' << s0)
	x4 := (1 << s0) + 'a'
	println(x3, x4)
	var i3 interface{} = x3
	_, ok3 := i3.(rune)
	println(ok3)

	println(('a' << s0) < 100)
	switch ('a' << s0) < 100 {
	case true:
		println("t")
	}

	x5 := ('a' << s0) * 2
	x6 := (1 << s0) + (1 << 40)
	println(x5, x6)

	var x7 interface{} = 1 + ('a' << s0)
	_, ok7 := x7.(rune)
	println(ok7)
	show(1 + ('a' << s0))
	_, ok8 := interface{}('a' << s0).(rune)
	println(ok8)
	m := map[interface{}]int{}
	m[1+('a'<<s0)] = 1
	_, ok9 := m[rune(98)]
	println(ok9)

	// Typed contexts.
	var i8 int8 = 1 << s0
	var u16 uint16 = 1 << s
	var f = 1 << s0
	println(i8, u16, f)

	// Rune ranking (#6339).
	println(int64((('a' + 1) << s) + (('b' + 1) << s)))
	println((('a' + 1) << s0) + (1 << s0))
	r := 'a' + 1
	rs := []rune{r}
	println(len(rs), r)

	// Constant counts.
	var x8 int64 = 'a' << 40
	var x9 uint8 = 1 << 7
	const k = 1 << 100
	println(x8, x9, k>>90, 1<<10, 1024>>10, 0<<0, 'a'>>1, uint64(1<<63)>>63)
}

// GnoOutput:

// GnoError:
// main/shift_context.go:21:6-40: cannot use int32 as int64

// GoOutput:
// 837518622720
// 837518622720
// 420906795008
// 420906795008
// false
// 98 98
// true
// true
// t
// 194 1099511627777
// true
// true
// true
// true
// 1 0 1
// 846108557312
// 99
// 1 98
// 106652627894272 128 1024 1024 1 0 48 1

// Verdict: KnownIssue run: #6393 non-constant shift operands are typed before their context is known (rune ranking: #6339)
