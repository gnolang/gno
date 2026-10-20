package main

// Non-constant shifts of untyped constants keep master's typing: these
// expressions are outside the exact constant folding, and each line must
// print what it printed before. Some of them differ from gc, which types a
// non-constant shift from its context (see the non-constant shift issue);
// the comments give gc's result where it differs.

var s0, s32 uint = 0, 32

const big2 = 'Ā' * 'Ā' * 'Ā' * 'Ā' // 4294967296

func f(x interface{}) interface{} { return x }

func main() {
	var s uint

	s = s32
	println(int64(('a' << s) + ('b' << s)))             // gc: 837518622720
	println(int64((('a' + 1) << s) + (('b' + 1) << s))) // also gc
	println(int64((('a' + 1) << s) + (1 << s)))         // also gc

	s = s0
	x1 := 1 + ('a' << s) // gc: 98
	x2 := (1 << s) + 'a' // gc: 98
	x3 := (('a' + 1) << s) + (1 << s)
	println(x1, x2, x3)
	println(('a' << s) < 100) // gc: true
	switch ('a' << s) < 100 {
	case true:
		println("gc only")
	}
	var i1 interface{} = 1 + ('a' << s)
	_, ok := i1.(rune) // gc: true
	println(ok, f(1+('a'<<s)))
	var i2 interface{} = ('a' << s) + (1 << 40) // gc: rejected
	println(i2, ('a'<<s)+(1<<40))               // gc: rejected

	u1 := +('a' << s)
	u2 := +(1 << s)
	var y1 int64 = u1           // gc: rejected
	var y2 int64 = u2           // gc: rejected
	println(y1, y2, u1+(1<<40)) // gc: rejected

	switch +('a' << s) {
	case 'a':
		println("+ match")
	}
	switch ('a' << s) + 'a' {
	case 'a' + 'a':
		println("compound match")
	}

	// Cancelled constants: big2*0 is 0.
	const c = big2*0 + 'a'
	const z = big2 * 0
	s = s32
	println(int64((c << s) + (c << s)))
	s = s0
	println(int64((z << s) + (1 << s)))
}

// Output:
// 0
// 846108557312
// 425201762304
// 1 1 99
// false
// false 98
// 97 97
// 97 1 97
// + match
// compound match
// 833223655424
// 1

// TypeCheckError:
// main/rune16.gno:36:37: 1 << 40 (untyped rune constant 1099511627776) overflows rune; main/rune16.gno:36:36: (1 << 40) (untyped rune constant 1099511627776) overflows rune; main/rune16.gno:37:24: 1 << 40 (untyped rune constant 1099511627776) overflows rune; main/rune16.gno:37:23: (1 << 40) (untyped rune constant 1099511627776) overflows rune; main/rune16.gno:41:17: cannot use u1 (variable of type rune) as int64 value in variable declaration; main/rune16.gno:42:17: cannot use u2 (variable of type int) as int64 value in variable declaration; main/rune16.gno:43:21: (1 << 40) (untyped int constant 1099511627776) overflows rune
