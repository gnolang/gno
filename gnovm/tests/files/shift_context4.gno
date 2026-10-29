package main

// Typed slots other than assignment: return, index, map key, call argument,
// composite element.
func f(v int64) int64 { return v }

func g(s uint) int64 { return ('a' << s) + ('b' << s) }

func h(s uint) rune { return 1 + ('a' << s) }

func main() {
	var s uint = 32
	println(f(('a' << s) + ('b' << s)))
	println(g(s))
	println(h(0))
	arr := [5]int{0, 1, 2, 3, 4}
	s = 1
	println(arr[(1 << s) + 1])
	m := map[int64]string{('a' << s) + ('b' << s): "x"}
	println(m[390])
	sl := []int64{('a' << s) + (1 << s), (1 << s) << s}
	println(sl[0], sl[1])
	type S struct{ v int64 }
	st := S{v: ('a' << s) + (1 << 40)}
	println(st.v)
}

// Output:
// 837518622720
// 837518622720
// 98
// 3
// x
// 196 4
// 1099511627970
