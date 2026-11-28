package main

// Each iteration of a three-clause loop gets its own copy of an array or
// struct declared in the init, as in Go 1.22, when a pointer or a closure
// keeps the variable past its iteration.

type S struct{ N int }

func main() {
	var ps []*[2]int
	for a := [2]int{}; a[0] < 3; a[0]++ {
		ps = append(ps, &a)
	}
	for _, p := range ps {
		println(p[0])
	}
	var fs []func() int
	for s := (S{}); s.N < 3; s.N++ {
		fs = append(fs, func() int { return s.N })
	}
	for _, f := range fs {
		println(f())
	}
	var is []func() int
	for i := 0; i < 3; i++ {
		is = append(is, func() int { return i })
	}
	for _, f := range is {
		println(f())
	}
}

// Output:
// 0
// 1
// 2
// 0
// 1
// 2
// 0
// 1
// 2
