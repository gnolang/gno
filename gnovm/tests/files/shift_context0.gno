package main

// Typed contexts: the shift operand takes the type of the whole expression.
func main() {
	var s uint = 32
	println(int64(('a' << s) + ('b' << s)))
	var x int64 = ('a' << s) + ('b' << s)
	println(x)
	println(int64(('a' << s) + (1 << s)))
	var y int64 = ('a' << s) + (1 << s)
	println(y)
	println(int64((('a' + 1) << s) + (('b' + 1) << s)))
	var z int64 = (('a' + 1) << s) + (1 << s)
	println(z)
	var w int64 = ('a' << s) + (1 << 40)
	println(w)
	var u uint64 = (1 << s) * 3
	println(u)
	var f = int64(-(1 << s))
	println(f)
	var g int64 = ((1 << s) << s) + 1
	println(g)
	x = (1 << s) + 'a'
	println(x)
	x += ('a' << s) - 1
	println(x)
}

// Output:
// 837518622720
// 837518622720
// 420906795008
// 420906795008
// 846108557312
// 425201762304
// 1516123455488
// 12884901888
// -4294967296
// 1
// 4294967393
// 420906795104
