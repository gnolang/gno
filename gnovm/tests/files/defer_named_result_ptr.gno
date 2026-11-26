package main

type S struct{ X int }

// p points into the named result; return s must keep the result s itself.
func f() (s S) {
	p := &s.X
	defer func() { *p = 2 }()
	return s
}

// sl aliases the named result array.
func g() (a [1]int) {
	sl := a[:]
	defer func() { sl[0] = 7 }()
	return a
}

func main() {
	println(f().X, g()[0])
}

// Output:
// 2 7
