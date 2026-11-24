package main

// Each iteration of a three-clause loop gets its own copy of an array or
// struct declared in the init, as in Go 1.22, when a slice of it or a pointer
// method value keeps the variable past its iteration. Slicing an array takes
// its address, as &a does.

type S struct{ buf [1]int }

type A [1]int

func (a *A) Get() int { return a[0] }

func main() {
	var ss [][]int
	for a := [1]int{}; a[0] < 3; a[0]++ {
		ss = append(ss, a[:])
	}
	for s := (S{}); s.buf[0] < 3; s.buf[0]++ {
		ss = append(ss, s.buf[:])
	}
	for a := [1][1]int{}; a[0][0] < 3; a[0][0]++ {
		ss = append(ss, a[0][:])
	}
	// A slice-typed init variable shares its backing array, as in Go.
	for s := []int{0}; s[0] < 3; s[0]++ {
		ss = append(ss, s[:])
	}
	for _, s := range ss {
		print(s[0])
	}
	println()

	var fs []func() int
	for a := (A{}); a[0] < 3; a[0]++ {
		fs = append(fs, a.Get)
	}
	for _, f := range fs {
		print(f())
	}
	println()
}

// Output:
// 012012012333
// 012
