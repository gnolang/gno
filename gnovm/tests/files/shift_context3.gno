package main

import "fmt"

// Interface contexts take the kind's default type.
func show(v interface{}) { fmt.Printf("%v %T\n", v, v) }

func main() {
	var s uint
	var x interface{} = 1 + ('a' << s)
	_, ok := x.(rune)
	println(ok)
	show(x)
	show((1 << s) + 'a')
	show(interface{}(('a' << s) * 2))
	show(interface{}((1 << s) + (1 << 40)))
	m := map[interface{}]string{}
	m[('a' << s) + 1] = "rune"
	m[(1 << s) + 1] = "int"
	println(m[rune(98)], m[int(2)], len(m))
	s = 32
	show(('a' << s) + ('b' << s))
	show(int64(('a' << s) + ('b' << s)))
	var y interface{} = ('a' << s) + (1 << s)
	show(y)
}

// Output:
// true
// 98 int32
// 98 int32
// 194 int32
// 1099511627777 int
// rune int 2
// 0 int32
// 837518622720 int64
// 0 int32
